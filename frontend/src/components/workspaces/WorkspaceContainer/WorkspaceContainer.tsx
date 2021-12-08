import { PlusOutlined, UserSwitchOutlined } from '@ant-design/icons';
import { Modal, Tooltip } from 'antd';
import Button from 'antd-button-color';
import { FC, useContext, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import {
  EnvironmentType,
  ImagesQuery,
  useCreateTemplateMutation,
  useImagesQuery,
} from '../../../generated-types';
import { Workspace } from '../../../utils';
import UserListLogic from '../../accountPage/UserListLogic/UserListLogic';
import Box from '../../common/Box';
import ModalCreateTemplate from '../ModalCreateTemplate';
import { Image, Template } from '../ModalCreateTemplate/ModalCreateTemplate';
import { TemplatesTableLogic } from '../Templates/TemplatesTableLogic';

export interface IWorkspaceContainerProps {
  tenantNamespace: string;
  workspace: Workspace;
}

const getImages = (dataImages: ImagesQuery) => {
  let images: Image[] = [];
  dataImages?.imageList?.images?.forEach(i => {
    const registry = i?.spec?.registryName!;
    const imagesRaw = i?.spec?.images;
    imagesRaw?.forEach(ir => {
      let versionsInImageName: Image[];
      if (registry === 'registry.internal.crownlabs.polito.it') {
        const latestVersion = `${ir?.name!}:${
          ir?.versions?.sort().reverse()[0]
        }`;
        versionsInImageName = [
          {
            name: latestVersion,
            vmorcontainer: ['VM'],
            registry: registry!,
          },
        ];
      } else {
        versionsInImageName = ir?.versions?.map(v => {
          return {
            name: `${ir?.name!}:${v}`,
            vmorcontainer: ['Container'],
            registry: registry!,
          };
        })!;
      }
      images = [...images, ...versionsInImageName!];
    });
  });
  return images;
};

const WorkspaceContainer: FC<IWorkspaceContainerProps> = ({ ...props }) => {
  const [showUserListModal, setShowUserListModal] = useState<boolean>(false);

  const {
    tenantNamespace,
    workspace: {
      role,
      name: workspaceName,
      namespace: workspaceNamespace,
      prettyName: workspacePrettyName,
    },
  } = props;

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [createTemplateMutation, { loading }] = useCreateTemplateMutation({
    onError: apolloErrorCatcher,
  });

  const [show, setShow] = useState(false);

  const { data: dataImages, refetch: refetchImages } = useImagesQuery({
    variables: {},
    onError: apolloErrorCatcher,
  });

  const submitHandler = (t: Template) =>
    createTemplateMutation({
      variables: {
        workspaceId: workspaceName,
        workspaceNamespace: workspaceNamespace,
        templateId: `${workspaceName}-`,
        templateName: t.name?.trim()!,
        descriptionTemplate: t.name?.trim()!,
        image: t.registry
          ? `${t.registry}/${t.image}`.trim()!
          : `${t.image}`.trim()!,
        guiEnabled: t.gui,
        persistent: t.diskMode,
        environmentType:
          t.vmorcontainer === EnvironmentType.Container
            ? EnvironmentType.Container
            : EnvironmentType.VirtualMachine,
        resources: {
          cpu: t.cpu,
          memory: `${t.ram * 1000}M`,
          disk: t.disk ? `${t.disk * 1000}M` : undefined,
          reservedCPUPercentage: 50,
        },
      },
    });

  return (
    <>
      <ModalCreateTemplate
        workspaceNamespace={workspaceNamespace}
        cpuInterval={{ max: 4, min: 1 }}
        ramInterval={{ max: 8, min: 1 }}
        diskInterval={{ max: 50, min: 10 }}
        setShow={setShow}
        show={show}
        images={getImages(dataImages!)}
        submitHandler={submitHandler}
        loading={loading}
      />
      <Box
        header={{
          size: 'large',
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>{workspacePrettyName}</b>
              </p>
            </div>
          ),
          left: role === 'manager' && (
            <div className="h-full flex justify-center items-center pl-10">
              <Tooltip title="Manage users">
                <Button
                  type="primary"
                  shape="circle"
                  size="large"
                  icon={<UserSwitchOutlined />}
                  onClick={() => setShowUserListModal(true)}
                />
              </Tooltip>
            </div>
          ),
          right: role === 'manager' && (
            <div className="h-full flex justify-center items-center pr-10">
              <Tooltip title="Create template">
                <Button
                  onClick={() => {
                    refetchImages();
                    setShow(true);
                  }}
                  type="lightdark"
                  shape="circle"
                  size="large"
                  icon={<PlusOutlined />}
                />
              </Tooltip>
            </div>
          ),
        }}
      >
        <TemplatesTableLogic
          tenantNamespace={tenantNamespace}
          role={role}
          workspaceNamespace={workspaceNamespace}
          workspaceName={workspaceName}
        />
        <Modal
          destroyOnClose={true}
          title={`Users in ${workspacePrettyName} `}
          width="800px"
          visible={showUserListModal}
          footer={null}
          onCancel={() => setShowUserListModal(false)}
        >
          <UserListLogic
            workspaceName={workspaceName}
            workspaceNamespace={workspaceNamespace}
          />
        </Modal>
      </Box>
    </>
  );
};

export default WorkspaceContainer;
