/* eslint-disable react/no-multi-comp */
import { Story } from '@storybook/react';
import { Title, Description, Stories } from '@storybook/addon-docs/blocks';
import RowHeading, { IRowHeadingProps } from './RowHeading';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/activePage/RowHeading',
  component: RowHeading,
  argTypes: { destroyAll: { action: 'clicked' } },
  parameters: {
    docs: {
      page: () => {
        <>
          <Title />
          <Description />
        </>;
      },
    },
  },
};

const defaultArgs: someKeysOf<IRowHeadingProps> = {
  text: 'Example',
  nActive: 1,
  newTempl: false,
};

const Template: Story<IRowHeadingProps> = args => <RowHeading {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    page: () => {
      return (
        <>
          <Title>Active Templates Accordion</Title>
          <Description>
            Heading for each Worksapce or Template. It displays an avatar with
            the number of templates or instances contained, a badge in case of
            newly created elemets and a Destroy All button. It has a responsive
            behavior as the avatar and the button collapse below the lg
            breakpoint.
          </Description>
          <Stories includePrimary />
        </>
      );
    },
  },
};

export const withBadge = Template.bind({});

withBadge.args = { ...defaultArgs, newTempl: true };
withBadge.parameters = {
  docs: {
    page: () => {
      return (
        <>
          <Title>Template header with badge</Title>
          <Description>
            A simple badge on the avatar highlights a recently created entry in
            the Table
          </Description>
          <Stories />
        </>
      );
    },
  },
};
