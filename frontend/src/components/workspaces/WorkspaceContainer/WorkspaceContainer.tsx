import { PlusOutlined, UserSwitchOutlined } from '@ant-design/icons';
import { Badge, Modal, Tooltip } from 'antd';
import { Button } from 'antd';
import type { FC } from 'react';
import { useContext, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import type { ImagesQuery } from '../../../generated-types';
import {
  EnvironmentType,
  useCreateTemplateMutation,
  useImagesQuery,
} from '../../../generated-types';
import type { Workspace } from '../../../utils';
import { JSONDeepCopy, WorkspaceRole } from '../../../utils';
import UserListLogic from '../../accountPage/UserListLogic/UserListLogic';
import Box from '../../common/Box';
import ModalCreateTemplate from '../ModalCreateTemplate';
import type {
  Image,
  Template,
} from '../ModalCreateTemplate/ModalCreateTemplate';
import { TemplatesTableLogic } from '../Templates/TemplatesTableLogic';

export interface IWorkspaceContainerProps {
  tenantNamespace: string;
  workspace: Workspace;
}

const getImages = (dataImages: ImagesQuery) => {
  let images: Image[] = [];
  JSONDeepCopy(dataImages?.imageList?.images)?.forEach(i => {
    const registry = i?.spec?.registryName;
    const imagesRaw = i?.spec?.images;
    imagesRaw?.forEach(ir => {
      let versionsInImageName: Image[];
      if (registry === 'registry.internal.crownlabs.polito.it') {
        const latestVersion = `${ir?.name}:${
          ir?.versions?.sort().reverse()[0]
        }`;
        versionsInImageName = [
          {
            name: latestVersion,
            vmorcontainer: [EnvironmentType.VirtualMachine],
            registry: registry!,
          },
        ];
      } else {
        versionsInImageName =
          ir?.versions.map(v => {
            return {
              name: `${ir?.name}:${v}`,
              vmorcontainer: [EnvironmentType.Container],
              registry: registry || '',
            };
          }) || [];
      }
      images = [...images, ...versionsInImageName!];
    });
  });
  return images;
};

const WorkspaceContainer: FC<IWorkspaceContainerProps> = ({ ...props }) => {
  const [showUserListModal, setShowUserListModal] = useState<boolean>(false);

  const { tenantNamespace, workspace } = props;

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
        workspaceId: workspace.name,
        workspaceNamespace: workspace.namespace,
        templateId: `${workspace.name}-`,
        templateName: t.name?.trim() || '',
        descriptionTemplate: t.name?.trim() || '',
        image: t.registry
          ? `${t.registry}/${t.image}`.trim()!
          : `${t.image}`.trim()!,
        guiEnabled: t.gui,
        persistent: t.persistent,
        mountMyDriveVolume: t.mountMyDrive,
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
        sharedVolumeMounts: t.sharedVolumeMountInfos ?? [],
      },
    });

  return (
    <>
      <ModalCreateTemplate
        workspaceNamespace={workspace.namespace}
        cpuInterval={{ max: 8, min: 1 }}
        ramInterval={{ max: 32, min: 1 }}
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
                <b>{workspace.prettyName}</b>
              </p>
            </div>
          ),
          left: workspace.role === WorkspaceRole.manager && (
            <div className="h-full flex justify-center items-center pl-10">
              <Tooltip title="Manage users">
                <Button
                  type="primary"
                  shape="circle"
                  size="large"
                  icon={<UserSwitchOutlined />}
                  onClick={() => setShowUserListModal(true)}
                >
                  {workspace.waitingTenants && (
                    <Badge
                      count={workspace.waitingTenants}
                      color="yellow"
                      className="absolute -top-2.5 -right-2.5"
                    />
                  )}
                </Button>
              </Tooltip>
            </div>
          ),
          right: workspace.role === WorkspaceRole.manager && (
            <div className="h-full flex justify-center items-center pr-10">
              <Tooltip title="Create template">
                <Button
                  onClick={() => {
                    refetchImages();
                    setShow(true);
                  }}
                  type="primary"
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
          role={workspace.role}
          workspaceNamespace={workspace.namespace}
          workspaceName={workspace.name}
        />
        <Modal
          destroyOnHidden={true}
          title={`Users in ${workspace.prettyName} `}
          width="800px"
          open={showUserListModal}
          footer={null}
          onCancel={() => setShowUserListModal(false)}
        >
          <UserListLogic workspace={workspace} />
        </Modal>
      </Box>
    </>
  );
};

export default WorkspaceContainer;
