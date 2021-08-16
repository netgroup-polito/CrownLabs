import TooltipButton, { ITooltipButtonProps } from './TooltipButton';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';
import {
  BarChartOutlined,
  DashboardOutlined,
  PieChartOutlined,
  QuestionOutlined,
} from '@ant-design/icons';
import Logo from '../Logo';

export default {
  title: 'Components/common/TooltipButton',
  component: TooltipButton,
  argTypes: {
    onClick: { table: { disable: true } },
    icon: { table: { disable: true } },
  },
  decorators: [
    (Story: Story) => (
      <div className="flex justify-center h-screen items-center">
        <Story />
      </div>
    ),
  ],
} as Meta;

const defaultArgs: someKeysOf<ITooltipButtonProps> = {
  TooltipButtonData: {
    type: 'primary',
    icon: <Logo widthPx={32} />,
  },
  onClick: () => null,
};

const Template: Story<ITooltipButtonProps> = args => (
  <TooltipButton {...args} />
);

export const Info = Template.bind({});

Info.args = defaultArgs;

export const Ticket = Template.bind({});

Ticket.args = {
  TooltipButtonData: {
    type: 'success',
    icon: (
      <QuestionOutlined
        style={{ fontSize: '22px' }}
        className="flex items-center justify-center "
      />
    ),
  },
  onClick: () => null,
};

export const Tooltip = Template.bind({});

Tooltip.args = {
  TooltipButtonData: {
    tooltipPlacement: 'right',
    tooltipTitle: 'This is a tooltip',
    type: 'warning',
    icon: (
      <DashboardOutlined
        style={{ fontSize: '22px' }}
        className="flex items-center justify-center "
      />
    ),
  },
  onClick: () => null,
};

export const ExternalLink = Template.bind({});

ExternalLink.args = {
  TooltipButtonData: {
    tooltipTitle: 'Click me to visit an external site',
    type: 'danger',
    icon: (
      <PieChartOutlined
        style={{ fontSize: '22px' }}
        className="flex items-center justify-center ml-0.5 "
      />
    ),
  },
  onClick: () => window.open('https://grafana.com/', '_blank'),
};

export const LotOfText = Template.bind({});

LotOfText.args = {
  TooltipButtonData: {
    tooltipTitle:
      'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.',
    tooltipPlacement: 'rightBottom',
    type: 'info',
    icon: (
      <BarChartOutlined
        style={{ fontSize: '22px' }}
        className="flex items-center justify-center "
      />
    ),
  },
  onClick: () => null,
};
