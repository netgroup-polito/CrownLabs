import { Button, Checkbox, Form, Row } from 'antd';
import { useContext, useState, type FC } from 'react';
import {
  TenantDocument,
  useApplyTenantJsonPatchJsonMutation,
  type TenantQuery,
} from '../../../generated-types';
import type { RuleRender, RuleObject } from 'antd/es/form';
import {
  convertToGiB,
  getCamelCaseKey,
  getOriginalK8sKey,
} from '../../../utils';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { CheckOutlined } from '@ant-design/icons';
import QuotaFields from '../../shared/QuotaFields';

export interface ITenantPersonalWorkspaceSettingsProps {
  tenant: TenantQuery;
}

interface QuotaFormData {
  enabled: boolean;
  cpu?: number;
  memory?: number;
  instances?: number;
  disk?: number;
  otherResources?: { key: string; value: number }[];
}

const TenantPersonalWorkspaceSettings: FC<
  ITenantPersonalWorkspaceSettingsProps
> = ({ tenant }) => {
  const [isEnabled, setIsEnabled] = useState(
    tenant.tenant?.spec?.personalWorkspace != null,
  );
  const [isSuccess, setIsSuccess] = useState(false);

  const [form] = Form.useForm<QuotaFormData>();

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [applyTenantJsonPatchJsonMutation, { loading }] =
    useApplyTenantJsonPatchJsonMutation({
      onError: apolloErrorCatcher,
    });

  const submitForm = async (data: QuotaFormData) => {
    setIsSuccess(false);

    const tenantId = tenant.tenant?.metadata?.name;
    if (!tenantId) {
      throw new Error('Tenant ID is missing');
    }

    let newQuota = null;
    if (data.enabled) {
      if (
        !data.cpu || data.cpu <= 0 ||
        !data.memory || data.memory <= 0 ||
        !data.instances || data.instances <= 0 ||
        data.disk == null || data.disk < 0
      ) {
        throw new Error('All quota fields must be provided when enabled');
      }

      // Convert form array in a map object
      const otherResourcesMap: { [key: string]: string } = {};
      data.otherResources?.forEach(res => {
        if (res.key && res.value != null && res.value > 0) {
          const k8sKey = getOriginalK8sKey(res.key);
          otherResourcesMap[k8sKey] = res.value.toString();
        }
      });

      newQuota = {
        cpu: data.cpu ?? 0,
        memory: `${data.memory?.toString() ?? '0'}Gi`,
        instances: data.instances ?? 0,
        disk: `${data.disk?.toString() ?? '0'}Gi`,
        otherResources: otherResourcesMap,
      };
    }

    const result = await applyTenantJsonPatchJsonMutation({
      variables: {
        tenantId: tenantId,
        patchJson: JSON.stringify([
          { op: 'add', path: '/spec/personalWorkspace', value: newQuota },
        ]),
        manager: 'frontend-tenant-personal-workspace',
      },
      // ensure Tenant query is refreshed so TenantContext and UI update
      refetchQueries: [
        { query: TenantDocument, variables: { tenantId: tenantId } },
      ],
      onError: apolloErrorCatcher,
    });

    if (result.errors == null) {
      setIsSuccess(true);
    }
  };

  // Strict validator for CPU, RAM, Instances (must be >= 1)
  const positiveNumberValidator: RuleRender = f => {
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

  // Flexible validator for Disk (allows 0)
  const nonNegativeNumberValidator: RuleRender = f => {
    if (f.getFieldValue('enabled')) {
      return {
        validator(_: RuleObject, value: number) {
          if (value >= 0) {
            return Promise.resolve();
          }
          return Promise.reject(new Error(`Value must be at least 0`));
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
    setIsSuccess(false);
    if (data.enabled !== undefined) setIsEnabled(data.enabled);
  };

  const currentWorkspace = tenant.tenant?.spec?.personalWorkspace;

  return (
    <Form
      form={form}
      labelCol={{ span: 6 }}
      wrapperCol={{ span: 18 }}
      onFinish={submitForm}
      onValuesChange={onValuesChange}
      initialValues={{
        enabled: tenant.tenant?.spec?.personalWorkspace != null,
        cpu: parseFloat(tenant.tenant?.spec?.personalWorkspace?.cpu ?? '0'),
        memory: convertToGiB(
          tenant.tenant?.spec?.personalWorkspace?.memory ?? '0GiB',
        ),
        instances: tenant.tenant?.spec?.personalWorkspace?.instances ?? 0,
        disk: convertToGiB(currentWorkspace?.disk ?? '0GiB'),
        // Convert otherResources object in an array
        otherResources: currentWorkspace?.otherResources
          ? Object.entries(currentWorkspace.otherResources).map(([k, v]) => ({
              key: getCamelCaseKey(k),
              value: parseFloat(v as string),
            }))
          : [],
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

      <QuotaFields
        disabled={!isEnabled}
        validateTrigger="onBlur"
        rules={{
          cpu: [positiveNumberValidator],
          memory: [positiveNumberValidator],
          instances: [positiveNumberValidator],
          disk: [nonNegativeNumberValidator],
          otherResources: [
            () => ({
              validator(_, value) {
                // If resource is enabled, values are validated
                if (!value || value.length === 0) return Promise.resolve();
                // Check non-negative values
                if (
                  value.every(
                    (res: { value?: number }) =>
                      res.value != null && res.value >= 0,
                  )
                ) {
                  return Promise.resolve();
                }
                return Promise.reject(
                  new Error('Resource amounts must be non-negative'),
                );
              },
            }),
          ],
        }}
        limits={{
          cpu: { min: 0 },
          memory: { min: 0 },
          instances: { min: 0 },
          disk: { min: 0 },
        }}
      />

      <Row justify="center">
        <Form.Item>
          <Button
            type="primary"
            color={isSuccess ? 'green' : 'primary'}
            variant="solid"
            htmlType="submit"
            loading={loading}
          >
            {isSuccess ? (
              <>
                <CheckOutlined /> Saved!
              </>
            ) : (
              'Save'
            )}
          </Button>
        </Form.Item>
      </Row>
    </Form>
  );
};

export default TenantPersonalWorkspaceSettings;
