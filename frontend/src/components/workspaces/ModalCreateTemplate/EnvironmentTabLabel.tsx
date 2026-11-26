import type { FC } from 'react';
import { EnvironmentType } from '../../../generated-types';
import { Form, Space } from 'antd';
import {
  DesktopOutlined,
  DockerOutlined,
  DeploymentUnitOutlined,
} from '@ant-design/icons';

type EnvironmentTabLabelProps = {
  envIndex: number;
};

export const EnvironmentTabLabel: FC<EnvironmentTabLabelProps> = ({
  envIndex,
}) => {
  const form = Form.useFormInstance();

  const name = Form.useWatch<string | undefined>([
    'environments',
    envIndex,
    'name',
  ]);

  const type = Form.useWatch<EnvironmentType>([
    'environments',
    envIndex,
    'environmentType',
  ]);

  Form.useWatch<string | undefined>(['environments', envIndex, 'image']);

  const nameErrors = form.getFieldError(['environments', envIndex, 'name']);
  const imageErrors = form.getFieldError(['environments', envIndex, 'image']);
  const workerErrors = form.getFieldError([
    'environments',
    envIndex,
    'cluster',
    'workersCount',
  ]);
  const hasErrors =
    nameErrors.length > 0 ||
    imageErrors.length > 0 ||
    workerErrors.length > 0 ||
    !name;

  return (
    <Space className={hasErrors ? 'text-red-500' : ''}>
      {type === EnvironmentType.Cluster ? (
        <DeploymentUnitOutlined />
      ) : type === EnvironmentType.Container ? (
        <DockerOutlined />
      ) : (
        <DesktopOutlined />
      )}

      <span>
        {!name ? 'Unknown' : name}
        {hasErrors && '*'}
      </span>
    </Space>
  );
};
