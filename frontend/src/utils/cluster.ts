import { EnvironmentType } from '../generated-types';

type EnvironmentKind = {
  environmentType?: EnvironmentType | null;
};

export const hasClusterEnvironment = (
  environments: readonly (EnvironmentKind | null | undefined)[],
): boolean =>
  environments.some(env => env?.environmentType === EnvironmentType.Cluster);

export const supportsEnvironmentSSH = (env: EnvironmentKind): boolean =>
  env.environmentType !== EnvironmentType.Cluster &&
  env.environmentType !== EnvironmentType.Container &&
  env.environmentType !== EnvironmentType.Standalone;

export const getEnvironmentWorkerCount = (
  env: EnvironmentKind & {
    cluster?: { workersCount?: number | null } | null;
  },
): number => {
  if (env.environmentType !== EnvironmentType.Cluster) return 1;

  const count = env.cluster?.workersCount;
  if (count == null || !Number.isInteger(count) || count < 1 || count > 10) {
    throw new Error(
      'Cluster environments require an integer worker count from 1 to 10.',
    );
  }
  return count;
};
