import { useContext, useState } from 'react';
import { Button, Input, Form, message } from 'antd';
import { useApplyTenantJsonPatchJsonMutation } from '../../../generated-types';
import type { TenantQuery } from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';

export interface ITenantSettingsProps {
  tenant: TenantQuery;
}

export default function TenantSettings({ tenant }: ITenantSettingsProps) {
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const tenantName = tenant.tenant?.metadata?.name || '';
  const currentLabels = tenant.tenant?.metadata?.labels || {};
  const currentOperatorSelector =
    currentLabels['crownlabsPolitoItOperatorSelector'] || '';

  const [operatorSelector, setOperatorSelector] = useState(
    currentOperatorSelector,
  );

  const [applyPatch, { loading }] = useApplyTenantJsonPatchJsonMutation({
    onCompleted: () => {
      message.success('Tenant settings updated successfully');
    },
    onError: apolloErrorCatcher,
  });

  const handleSave = () => {
    if (!tenantName) return;

    applyPatch({
      variables: {
        tenantId: tenantName,
        patchJson: JSON.stringify([
          {
            op: 'replace',
            path: '/metadata/labels/crownlabs.polito.it~1operator-selector',
            value: operatorSelector,
          },
        ]),
        manager: 'crownlabs-frontend',
      },
    });
  };

  return (
    <div className="p-4">
      <Form layout="vertical">
        <Form.Item label="Operator Selector">
          <Input
            value={operatorSelector}
            onChange={e => setOperatorSelector(e.target.value)}
            placeholder="Enter operator selector"
          />
        </Form.Item>
        <Button type="primary" onClick={handleSave} loading={loading}>
          Save
        </Button>
      </Form>
    </div>
  );
}
