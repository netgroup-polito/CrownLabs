import { FC, useState } from 'react';
import { UserSwitchOutlined, PlusOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';
import Box from '../../common/Box';
import { TemplatesTableLogic } from '../Templates/TemplatesTableLogic';
import { WorkspaceRole } from '../../../utils';
import {
  EnvironmentType,
  ImagesQuery,
  useCreateTemplateMutation,
  useImagesQuery,
} from '../../../generated-types';
import ModalCreateTemplate from '../ModalCreateTemplate';
import { Image, Template } from '../ModalCreateTemplate/ModalCreateTemplate';
import { Tooltip } from 'antd';

export interface IWorkspaceContainerProps {
  tenantNamespace: string;
  workspace: {
    id: number;
    title: string;
    role: WorkspaceRole;
    workspaceNamespace: string;
    workspaceName: string;
  };
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
  const {
    tenantNamespace,
    workspace: { role, title, workspaceNamespace, workspaceName },
  } = props;

  const [createTemplateMutation, { loading }] = useCreateTemplateMutation();

  const [show, setShow] = useState(false);

  const { data: dataImages, refetch: refetchImages } = useImagesQuery({
    variables: {},
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
                <b>{title}</b>
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
        />
      </Box>
    </>
  );
};

export default WorkspaceContainer;
