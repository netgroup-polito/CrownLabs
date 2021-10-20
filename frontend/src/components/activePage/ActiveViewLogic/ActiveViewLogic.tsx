import { FC } from 'react';
import { workspaces, templates } from '../tempData';
import ActiveView from '../ActiveView/ActiveView';
import { Template } from '../../../utils';

export interface IActiveViewLogicProps {}

const ActiveViewLogic: FC<IActiveViewLogicProps> = ({ ...props }) => {
  return (
    <ActiveView
      workspaces={workspaces}
      instances={templates
        .map((template: Template) => [...template.instances])
        .flat(1)}
      managerView={true}
    />
  );
};

export default ActiveViewLogic;
