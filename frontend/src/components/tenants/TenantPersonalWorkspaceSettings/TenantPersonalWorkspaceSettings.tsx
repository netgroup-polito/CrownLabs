import { Button, Checkbox, Form, InputNumber, Row } from 'antd';
import { useContext, useState, type FC } from 'react';
import {
  TenantDocument,
  useApplyTenantMutation,
  type TenantQuery,
} from '../../../generated-types';
import type { RuleRender, RuleObject } from 'antd/es/form';
import { convertToGB } from '../../../utils';
import { getTenantPatchJson } from '../../../graphql-components/utils';
import { ErrorContext } from '../../../errorHandling/ErrorContext';

export interface ITenantPersonalWorkspaceSettingsProps {
  tenant: TenantQuery;
}

interface QuotaFormData {
  enabled: boolean;
  cpu?: number;
  memory?: number;
  instances?: number;
}

const TenantPersonalWorkspaceSettings: FC<
  ITenantPersonalWorkspaceSettingsProps
> = ({ tenant }) => {
  const [isEnabled, setIsEnabled] = useState(
    tenant.tenant?.spec?.quota != null,
  );

  const [form] = Form.useForm<QuotaFormData>();

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [applyTenantMutation] = useApplyTenantMutation({
    onError: apolloErrorCatcher,
  });

  const submitForm = async (data: QuotaFormData) => {
    const tenantId = tenant.tenant?.metadata?.name;
    if (!tenantId) {
      throw new Error('Tenant ID is missing');
    }

    let newQuota = null;
    if (data.enabled) {
      if (!data.cpu || !data.memory || !data.instances) {
        throw new Error('All quota fields must be provided when enabled');
      }

      newQuota = {
        cpu: data.cpu?.toString() ?? '0',
        memory: `${data.memory?.toString() ?? '0'}Gi`,
        instances: data.instances ?? 0,
      };
    }

    await applyTenantMutation({
      variables: {
        tenantId: tenantId,
        patchJson: getTenantPatchJson({
          quota: newQuota,
        }),
        manager: 'frontend-tenant-personal-workspace',
      },
      // ensure Tenant query is refreshed so TenantContext and UI update
      refetchQueries: [
        { query: TenantDocument, variables: { tenantId: tenantId } },
      ],
      onError: apolloErrorCatcher,
    });
  };

  const numberValidator: RuleRender = f => {
    if (f.getFieldValue('enabled')) {
      return {
        validator(_: RuleObject, value: number) {
          if (value >= 1) {
            return Promise.resolve();
          }
          return Promise.reject(new Error(`Value must be at least 1`));
        },
      };
    } else {
      return {
        validator(_: RuleObject, _value: number) {
          return Promise.resolve();
        },
      };
    }
  };

  const onValuesChange = (data: QuotaFormData) => {
    if (data.enabled !== undefined) setIsEnabled(data.enabled);
  };

  return (
    <Form
      form={form}
      labelCol={{ span: 12 }}
      wrapperCol={{ span: 12 }}
      onFinish={submitForm}
      onValuesChange={onValuesChange}
      initialValues={{
        enabled: tenant.tenant?.spec?.quota != null,
        cpu: parseFloat(tenant.tenant?.spec?.quota?.cpu ?? '0'),
        memory: convertToGB(tenant.tenant?.spec?.quota?.memory ?? '0'),
        instances: tenant.tenant?.spec?.quota?.instances ?? 0,
      }}
    >
      <Form.Item
        name="enabled"
        valuePropName="checked"
        label="Enabled"
        validateTrigger="onBlur"
      >
        <Checkbox />
      </Form.Item>

      <Form.Item
        name="cpu"
        label="CPU"
        validateTrigger="onBlur"
        rules={[numberValidator]}
      >
        <InputNumber min={0} disabled={!isEnabled} />
      </Form.Item>

      <Form.Item
        name="memory"
        label="Memory (GB)"
        validateTrigger="onBlur"
        rules={[numberValidator]}
      >
        <InputNumber min={0} disabled={!isEnabled} />
      </Form.Item>

      <Form.Item
        name="instances"
        label="Instances"
        validateTrigger="onBlur"
        rules={[numberValidator]}
      >
        <InputNumber min={0} disabled={!isEnabled} />
      </Form.Item>

      <Row justify="center">
        <Form.Item>
          <Button type="primary" htmlType="submit">
            Save
          </Button>
        </Form.Item>
      </Row>
    </Form>
  );
};

export default TenantPersonalWorkspaceSettings;
