import type { ApolloError } from '@apollo/client';
import { Alert, Button, Form, Input } from 'antd';
import { type FC, useEffect, useMemo, useState } from 'react';
import {
  CommandPolicy,
  useCreateCrownJobMutation,
  useWorkspaceJobLimitQuery,
} from '../../../../generated-types';
import type { Instance, InstanceJobTemplate } from '../../../../utils';
import {
  apiErrorTexts,
  COMMAND_MAX_LENGTH,
  commandToText,
  CROWNJOB_NAME_MAX_LENGTH,
  CROWNJOB_NAME_PATTERN,
  effectiveRunLimit,
  jobCommand,
  MAX_RUN_TIME_PATTERN,
  nextJobName,
} from './crownJobUtils';

export interface IRunJobFormProps {
  instance: Instance;
  job: InstanceJobTemplate;
  usedNames: ReadonlySet<string>;
  onCreated: () => Promise<unknown>;
}

interface RunJobFormValues {
  name: string;
  command?: string;
  maxRunTime?: string;
}

const RunJobForm: FC<IRunJobFormProps> = ({
  instance,
  job,
  usedNames,
  onCreated,
}) => {
  const [form] = Form.useForm<RunJobFormValues>();
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState<string[]>([]);

  // An error rejects the promise and is shown here, not in the global popup.
  const [createCrownJob] = useCreateCrownJobMutation();

  const { data: workspaceData } = useWorkspaceJobLimitQuery({
    variables: { name: instance.workspaceName },
    skip: !instance.workspaceName,
  });
  const limit = effectiveRunLimit(
    job.maxRunTime,
    workspaceData?.workspace?.spec?.queueLimits?.maxRunTime,
  );

  const suggestedName = useMemo(
    () => nextJobName(instance.name, usedNames),
    [instance.name, usedNames],
  );

  // setFieldValue, unlike setFieldsValue, does not mark the field as touched:
  // the suggestion keeps following the list until the user edits the name.
  useEffect(() => {
    if (!form.isFieldTouched('name')) {
      form.setFieldValue('name', suggestedName);
    }
  }, [form, suggestedName]);

  const onFinish = async (values: RunJobFormValues) => {
    setErrors([]);
    setSubmitting(true);
    try {
      await createCrownJob({
        variables: {
          tenantNamespace: instance.tenantNamespace,
          name: values.name,
          instanceName: instance.name,
          command: jobCommand(job.commandPolicy, values.command),
          maxRunTime: values.maxRunTime?.trim() || undefined,
        },
      });
    } catch (error) {
      setErrors(apiErrorTexts(error as ApolloError));
      setSubmitting(false);
      return;
    }
    // The job exists: a failed refresh is reported by the list, not here.
    await onCreated().catch(() => null);
    form.resetFields(['name']);
    setSubmitting(false);
  };

  return (
    <Form
      form={form}
      layout="vertical"
      requiredMark={false}
      initialValues={{
        name: suggestedName,
        command: commandToText(job.command),
      }}
      onFinish={onFinish}
    >
      {job.commandPolicy === CommandPolicy.Editable && (
        <Form.Item
          name="command"
          label="Command"
          extra="runs with sh -c"
          rules={[
            {
              required: true,
              whitespace: true,
              message: 'Write the command to run',
            },
          ]}
        >
          <Input.TextArea
            autoSize={{ minRows: 2, maxRows: 6 }}
            maxLength={COMMAND_MAX_LENGTH}
            placeholder="python train.py"
          />
        </Form.Item>
      )}
      {job.commandPolicy === CommandPolicy.Fixed && (
        <Form.Item label="Command" extra="set by the Template">
          <Input.TextArea
            readOnly
            autoSize={{ minRows: 1, maxRows: 6 }}
            value={commandToText(job.command)}
          />
        </Form.Item>
      )}
      <div className="flex flex-col sm:flex-row gap-x-4">
        <Form.Item
          className="flex-1"
          name="name"
          label="Job name"
          rules={[
            { required: true, message: 'Give the job a name' },
            {
              max: CROWNJOB_NAME_MAX_LENGTH,
              message: `At most ${CROWNJOB_NAME_MAX_LENGTH} characters`,
            },
            {
              pattern: CROWNJOB_NAME_PATTERN,
              message:
                'Lowercase letters, digits, "-" and ".", starting and ending with a letter or a digit',
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          className="flex-1"
          name="maxRunTime"
          label="Time limit (optional)"
          extra={
            limit
              ? `Limit: ${limit.value} (${limit.source}), set-up included`
              : "Limit: the operator's default, set-up included"
          }
          rules={[
            {
              pattern: MAX_RUN_TIME_PATTERN,
              message: 'A number followed by m, h or d (e.g., 30m, 6h, 2d)',
            },
          ]}
        >
          <Input placeholder="e.g., 30m" allowClear />
        </Form.Item>
      </div>
      {errors.length > 0 && (
        <Alert
          className="mb-4"
          type="error"
          showIcon
          message={errors.map(text => (
            <div key={text}>{text}</div>
          ))}
        />
      )}
      <div className="flex justify-end">
        <Button
          type="primary"
          shape="round"
          htmlType="submit"
          loading={submitting}
        >
          Create
        </Button>
      </div>
    </Form>
  );
};

export default RunJobForm;
