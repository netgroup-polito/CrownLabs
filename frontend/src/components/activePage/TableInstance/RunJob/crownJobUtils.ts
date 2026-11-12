import type { ApolloError } from '@apollo/client';
import {
  CommandPolicy,
  EnvironmentType,
  Phase as CrownJobPhase,
  type CrownJobsQuery,
} from '../../../../generated-types';
import type { Instance } from '../../../../utils';

type CrownJobList = NonNullable<CrownJobsQuery['crownJobList']>;
export type CrownJob = NonNullable<NonNullable<CrownJobList['jobs']>[number]>;

// A CrownJob has the rules of a Kubernetes object name, and at most 63
// characters because its Job takes the same name (forge.CrownJobMaxNameLength).
export const CROWNJOB_NAME_MAX_LENGTH = 63;
export const CROWNJOB_NAME_PATTERN =
  /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;

// The pattern of spec.maxRunTime in the CRD.
export const MAX_RUN_TIME_PATTERN = /^[1-9][0-9]{0,3}[mhd]$/;

// The items:MaxLength of spec.command in the CRD.
export const COMMAND_MAX_LENGTH = 8192;

// A job needs a persistent container: the Template's CEL rule says so too.
export const isRunJobAvailable = (
  instance: Pick<Instance, 'job' | 'environments'>,
): boolean => {
  const { job, environments } = instance;
  const environment = environments?.find(
    env => env.name === job?.environmentName,
  );
  return (
    Boolean(job) &&
    Boolean(environment?.persistent) &&
    (environment?.environmentType === EnvironmentType.Container ||
      environment?.environmentType === EnvironmentType.Standalone)
  );
};

const shellQuote = (arg: string) =>
  /^[A-Za-z0-9_@%+=:,./-]+$/.test(arg)
    ? arg
    : `'${arg.replace(/'/g, `'\\''`)}'`;

// The text of a Template's command: the list of arguments as one line, or the
// script itself if the list is already sh -c <script>.
export const commandToText = (command: readonly string[]): string => {
  const [shell, flag, script] = command;
  if (
    command.length === 3 &&
    ['sh', 'bash', '/bin/sh', '/bin/bash'].includes(shell) &&
    flag === '-c'
  ) {
    return script;
  }
  return command.map(shellQuote).join(' ');
};

// spec.command: the operator refuses a command unless the policy is Editable.
export const jobCommand = (
  policy: CommandPolicy,
  text: string | undefined,
): string[] | undefined =>
  policy === CommandPolicy.Editable && text?.trim()
    ? ['sh', '-c', text.trim()]
    : undefined;

const toMinutes = (value: string): number | null => {
  const match = /^(\d+)([mhd])$/.exec(value);
  if (!match) return null;
  return (
    Number(match[1]) * { m: 1, h: 60, d: 1440 }[match[2] as 'm' | 'h' | 'd']
  );
};

export type JobRunLimit = { value: string; source: 'Template' | 'Workspace' };

// The same rule as the operator (forge.CrownJobMaxRunTime): the smaller of the
// two. With neither, the default of the operator applies.
export const effectiveRunLimit = (
  templateValue?: string | null,
  workspaceValue?: string | null,
): JobRunLimit | null => {
  const template: JobRunLimit | null = templateValue
    ? { value: templateValue, source: 'Template' }
    : null;
  const workspace: JobRunLimit | null = workspaceValue
    ? { value: workspaceValue, source: 'Workspace' }
    : null;
  if (!template || !workspace) return template ?? workspace;
  const templateMinutes = toMinutes(template.value);
  const workspaceMinutes = toMinutes(workspace.value);
  if (templateMinutes === null || workspaceMinutes === null) return template;
  return workspaceMinutes < templateMinutes ? workspace : template;
};

// The first free name <instance>-<n>: names are unique in the namespace.
export const nextJobName = (
  instanceName: string,
  usedNames: ReadonlySet<string>,
): string => {
  for (let n = 1; ; n++) {
    const suffix = `-${n}`;
    const base = instanceName
      .slice(0, CROWNJOB_NAME_MAX_LENGTH - suffix.length)
      .replace(/[-.]+$/, '');
    if (!usedNames.has(base + suffix)) return base + suffix;
  }
};

export const isJobEnded = (job: CrownJob): boolean =>
  job.status?.phase === CrownJobPhase.Succeeded ||
  job.status?.phase === CrownJobPhase.Failed;

const createdAt = (job: CrownJob): number =>
  new Date(job.metadata?.creationTimestamp ?? 0).getTime();

export const instanceJobs = (
  jobs: readonly (CrownJob | null)[],
  instanceName: string,
): CrownJob[] =>
  jobs
    .filter(
      (job): job is CrownJob => job?.spec.instanceRef.name === instanceName,
    )
    .sort(
      (a, b) =>
        createdAt(b) - createdAt(a) ||
        (a.metadata?.name ?? '').localeCompare(b.metadata?.name ?? ''),
    );

export const PHASE_COLOR: Record<CrownJobPhase, string> = {
  [CrownJobPhase.Waiting]: 'gold',
  [CrownJobPhase.Starting]: 'cyan',
  [CrownJobPhase.Running]: 'blue',
  [CrownJobPhase.Succeeded]: 'green',
  [CrownJobPhase.Failed]: 'red',
};

export type JobStage = {
  title: string;
  status: 'wait' | 'process' | 'finish' | 'error';
};

const STAGES = [
  CrownJobPhase.Waiting,
  CrownJobPhase.Starting,
  CrownJobPhase.Running,
];

export const jobStages = (phase?: CrownJobPhase | null): JobStage[] => {
  if (phase === CrownJobPhase.Succeeded) {
    return [...STAGES, CrownJobPhase.Succeeded].map(title => ({
      title,
      status: 'finish',
    }));
  }
  if (phase === CrownJobPhase.Failed) {
    // A failed job may never have started, so the stages before it stay off.
    return [
      ...STAGES.map(title => ({ title, status: 'wait' as const })),
      { title: CrownJobPhase.Failed, status: 'error' },
    ];
  }
  const current = Math.max(STAGES.indexOf(phase as CrownJobPhase), 0);
  return [
    ...STAGES.map((title, index) => ({
      title,
      status:
        index < current
          ? ('finish' as const)
          : index === current
            ? ('process' as const)
            : ('wait' as const),
    })),
    { title: 'Ended', status: 'wait' },
  ];
};

export const formatJobTime = (timestamp?: string | null): string => {
  if (!timestamp) return '-';
  const date = new Date(timestamp);
  return Number.isNaN(date.getTime()) ? timestamp : date.toLocaleString();
};

// The words of the API: qlkube puts the Kubernetes Status in responseBody, and
// its own message says only which operation failed.
export const apiErrorTexts = (error: ApolloError): string[] => {
  const texts = (error.graphQLErrors ?? []).map(graphQLError => {
    const body = graphQLError.extensions?.responseBody as
      | { message?: unknown }
      | string
      | undefined;
    if (typeof body === 'string' && body) return body;
    if (body && typeof body === 'object' && typeof body.message === 'string') {
      return body.message;
    }
    return graphQLError.message;
  });
  return texts.length ? texts : [error.message];
};
