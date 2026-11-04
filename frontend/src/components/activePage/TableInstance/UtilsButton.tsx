import { InfoOutlined } from '@ant-design/icons';
import { Button, Popover, Typography } from 'antd';
import { useMemo, type FC } from 'react';
const { Text } = Typography;

export type UtilsButtonProps = {
  user?: string | false;
  instanceName?: string | false;
  workspace?: string | false;
  template?: string | false;
  age?: string | false;
  lastAccess?: string | false;
};

const UtilsButton: FC<UtilsButtonProps> = props => {
  const { age, instanceName, lastAccess, template, user, workspace } = props;

  const content = useMemo(() => {
    const labels = [
      { label: 'User', val: user },
      { label: 'Instance Name', val: instanceName },
      { label: 'Workspace', val: workspace },
      { label: 'Template', val: template },
      { label: 'Age', val: age },
      { label: 'Last access', val: lastAccess },
    ];

    return (
      <>
        {labels.map(
          ({ label, val }) =>
            val && (
              <p className="m-0" key={label}>
                <strong>{label}: </strong>
                <Text italic>{val}</Text>
              </p>
            ),
        )}
      </>
    );
  }, [age, instanceName, lastAccess, template, user, workspace]);

  return (
    <Popover placement="top" content={content} trigger="click">
      <Button shape="circle" className="mr-3">
        <InfoOutlined />
      </Button>
    </Popover>
  );
};

export default UtilsButton;
