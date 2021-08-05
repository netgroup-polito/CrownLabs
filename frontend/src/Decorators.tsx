//This component will be removed in the next weeks after a decorators refactoring
/* eslint-disable react/no-multi-comp */
import { useState } from 'react';
import { Layout, Skeleton } from 'antd';
import Button from 'antd-button-color';
import { StoryContext } from '@storybook/react/dist/ts3.9/client/preview/types';
import { Story } from '@storybook/react';
import { MemoryRouter, Route } from 'react-router-dom';
import ThemeContextProvider from './contexts/ThemeContext';

const ModalDecorator = (Story: Story, context: StoryContext) => {
  const [show, setShow] = useState(false);
  context.args.showModal = show;
  context.args.setShowModal = setShow;

  return (
    <Layout>
      <Layout.Content>
        <Story />
        <div className="m-4">
          <Skeleton />
        </div>
        <div className=" flex justify-center mt-12">
          <Button
            onClick={() => setShow(true)}
            type="primary"
            shape="round"
            size="large"
          >
            Open Modal
          </Button>
        </div>
      </Layout.Content>
    </Layout>
  );
};

const LogoDecorator = (Story: Story) => (
  <div className="flex h-full justify-center h-screen ">
    <Story />
  </div>
);

const NavbarDecorator = (Story: Story) => (
  <MemoryRouter>
    <ThemeContextProvider>
      <Route path="/">
        <Layout>
          <Story />
          <Layout.Content>
            <div className="m-4">
              <Skeleton />
              <Skeleton />
              <Skeleton />
            </div>
          </Layout.Content>
        </Layout>
      </Route>
    </ThemeContextProvider>
  </MemoryRouter>
);

export { ModalDecorator, NavbarDecorator, LogoDecorator };
