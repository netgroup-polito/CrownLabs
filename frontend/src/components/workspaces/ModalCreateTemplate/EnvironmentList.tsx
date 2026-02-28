import { Form, Tabs } from 'antd';
import { useState, type FC } from 'react';
import { EnvironmentTabLabel } from './EnvironmentTabLabel';
import { EnvironmentType } from '../../../generated-types';
import type { SharedVolume } from '../../../utils';
import { Environment } from './Environment';
import type { Resources, TemplateFormEnv, Image } from './types';

const getDefaultEnvironment = (envCount: number): TemplateFormEnv => {
  const name = `env-${envCount}`;

  return {
    name: name,
    image: '',
    registry: '',
    environmentType: EnvironmentType.VirtualMachine,
    persistent: false,
    gui: true,
    cpu: 1,
    ram: 1,
    disk: 0,
    reservedCpu: 50,
    sharedVolumeMounts: [],
    rewriteUrl: false,
  };
};

interface IEnvironmentLabelProps {
  availableImages: Image[];
  resources: Resources;
  sharedVolumes: SharedVolume[];
  isPersonal: boolean;
  setInfoNumberTemplate: React.Dispatch<React.SetStateAction<number>>;
}

export const EnvironmentList: FC<IEnvironmentLabelProps> = ({
  availableImages,
  resources,
  sharedVolumes,
  isPersonal,
  setInfoNumberTemplate
}) => {
  const form = Form.useFormInstance();
  const environments = Form.useWatch<TemplateFormEnv[] | undefined>(
    'environments',
  );

  const [activeTabItem, setActiveTabItem] = useState('0');

  const addEnv = () => {
    const envIndex = environments ? environments.length : 0;

    form.setFieldsValue({
      environments: [
        ...(environments || []),
        getDefaultEnvironment(envIndex + 1),
      ],
    });

    setActiveTabItem(envIndex.toString());
  };

  const removeEnv = (targetKey: string) => {
    if (!environments) return;

    const targetIndex = parseInt(targetKey);

    const filteredEnvironments = environments.filter(
      (_, envIdx) => targetIndex !== envIdx,
    );

    if (filteredEnvironments.length === 0) {
      form.setFieldsValue({ environments: [getDefaultEnvironment(1)] });
      setActiveTabItem('0');
      return;
    }

    form.setFieldsValue({ environments: filteredEnvironments });

    if (targetKey === activeTabItem) {
      if (targetIndex === 0) {
        setActiveTabItem('0');
      } else {
        setActiveTabItem((targetIndex - 1).toString());
      }
    }
  };

  const handleTabEdit = (targetKey: string, action: 'add' | 'remove') => {
    switch (action) {
      case 'add':
        addEnv();
        setInfoNumberTemplate((prev) => prev + 1);
        return;
      case 'remove':
        removeEnv(targetKey);
        setInfoNumberTemplate((prev) => prev == 1 ? prev : prev - 1);
        return;
    }
  };

  return (
    <>
      {/* <div className="mb-2">
        <Typography.Text strong>Virtual Machines / Containers</Typography.Text>
      </div> */}

      <Form.List name="environments">
        {(fields, _) => (
          <>
            <Tabs
              type="editable-card"
              activeKey={activeTabItem}
              items={fields.map(({ key, name, ...restField }) => ({
                key: key.toString(),
                label: <EnvironmentTabLabel envIndex={name} />,
                children: (
                  <Environment
                    restField={restField}
                    parentFormName={name}
                    availableImages={availableImages}
                    resources={resources}
                    sharedVolumes={sharedVolumes}
                    isPersonal={isPersonal}
                  />
                ),
              }))}
              onChange={setActiveTabItem}
              onEdit={(target, action) =>
                handleTabEdit(target.toString(), action)
              }
            />
          </>
        )}
      </Form.List>
    </>
  );
};
