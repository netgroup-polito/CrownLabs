import type { FC } from 'react';
import { useState, useContext, useEffect } from 'react';
import { Modal, Form, Input } from 'antd';
import { Button } from 'antd';
import type { CreateTemplateMutation } from '../../../generated-types';
import {
  EnvironmentType,
  useWorkspaceTemplatesQuery,
  useImagesQuery,
  useWorkspaceSharedVolumesQuery,
} from '../../../generated-types';
import type { ApolloError, FetchResult } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { makeGuiSharedVolume } from '../../../utilsLogic';
import type { SharedVolume } from '../../../utils';
import { EnvironmentList } from './EnvironmentList';
import type { Image, Interval, TemplateForm } from './types';
import {
  getDefaultTemplate,
  getImageLists,
  getImageNameNoVer,
  getImagesFromList,
  internalRegistry,
} from './utils';

export interface IModalCreateTemplateProps {
  workspaceNamespace: string;
  template?: TemplateForm;
  cpuInterval: Interval;
  ramInterval: Interval;
  diskInterval: Interval;
  show: boolean;
  setShow: (status: boolean) => void;
  submitHandler: (
    t: TemplateForm,
  ) => Promise<
    FetchResult<
      CreateTemplateMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  loading: boolean;
  isPersonal?: boolean;
}

const ModalCreateTemplate: FC<IModalCreateTemplateProps> = ({ ...props }) => {
  const {
    show,
    setShow,
    cpuInterval,
    ramInterval,
    diskInterval,
    template,
    submitHandler,
    loading,
    workspaceNamespace,
    isPersonal,
  } = props;

  const { apolloErrorCatcher } = useContext(ErrorContext);

  // Fetch all image lists
  const { data: dataImages } = useImagesQuery({
    variables: {},
    onError: apolloErrorCatcher,
  });

  const [form] = Form.useForm<TemplateForm>();

  const [sharedVolumes, setDataShVols] = useState<SharedVolume[]>([]);

  useWorkspaceSharedVolumesQuery({
    variables: { workspaceNamespace },
    onError: apolloErrorCatcher,
    onCompleted: data =>
      setDataShVols(
        data.sharedvolumeList?.sharedvolumes
          ?.map(sv => makeGuiSharedVolume(sv))
          .sort((a, b) =>
            (a.prettyName ?? '').localeCompare(b.prettyName ?? ''),
          ) ?? [],
      ),
    fetchPolicy: 'network-only',
  });

  const validateName = async (_: unknown, name: string) => {
    if (!dataFetchTemplates || loadingFetchTemplates || errorFetchTemplates) {
      throw new Error('Error fetching templates');
    }

    if (!dataFetchTemplates.templateList) return;

    const trimmedName = name.trim().toLowerCase();
    const duplicateIndex = dataFetchTemplates.templateList.templates.findIndex(
      t => {
        return t?.spec?.prettyName?.toLowerCase() === trimmedName;
      },
    );

    if (duplicateIndex !== -1) {
      throw new Error(`This name has already been used in this workspace`);
    }
  };

  const fullLayout = {
    wrapperCol: { offset: 0, span: 24 },
  };

  const closehandler = () => {
    setShow(false);
  };

  const {
    data: dataFetchTemplates,
    error: errorFetchTemplates,
    loading: loadingFetchTemplates,
  } = useWorkspaceTemplatesQuery({
    onError: error => {
      console.error(
        'ModalCreateTemplate useWorkspaceTemplatesQuery error:',
        error,
        'workspaceNamespace:',
        workspaceNamespace,
      );
      apolloErrorCatcher(error);
    },
    variables: { workspaceNamespace },
  });

  const [availableImages, setAvailableImages] = useState<Image[]>([]);

  useEffect(() => {
    if (!dataImages) {
      setAvailableImages([]);
      return;
    }

    const imageLists = getImageLists(dataImages);
    const internalImages = imageLists.find(
      list => list.registryName === internalRegistry,
    );

    if (!internalImages) {
      setAvailableImages([]);
      return;
    }

    setAvailableImages(getImagesFromList(internalImages));
  }, [dataImages]);

  // Determine the final image URL
  const parseImage = (envType: EnvironmentType, image: string): string => {
    if (envType === EnvironmentType.VirtualMachine) {
      // For VMs, use the selected image from internal registry
      const selectedImage = availableImages.find(
        i => getImageNameNoVer(i.name) === image,
      );

      if (selectedImage) {
        return `${internalRegistry}/${selectedImage.name}`;
      }
    }

    // For other types, use the external image
    let finalImage = image;
    // If it doesn't include a registry, default to internal registry
    if (finalImage && !finalImage.includes('/') && !finalImage.includes('.')) {
      finalImage = `${internalRegistry}/${finalImage}`;
    }

    return finalImage;
  };

  const handleFormFinish = async (template: TemplateForm) => {
    // Prepare the template (parse the image URLs)
    const parsedTemplate = {
      ...template,
      environments: template.environments.map(env => ({
        ...env,
        image: parseImage(env.environmentType, env.image),
      })),
    };

    try {
      await submitHandler(parsedTemplate);

      setShow(false);
      form.resetFields();
    } catch (error) {
      console.error('ModalCreateTemplate submitHandler error:', error);
      apolloErrorCatcher(error as ApolloError);
    }
  };

  const getInitialValues = (template?: TemplateForm) => {
    if (template) return template;

    return getDefaultTemplate({
      cpu: cpuInterval,
      ram: ramInterval,
      disk: diskInterval,
    });
  };

  const handleFormSubmit = async () => {
    try {
      await form.validateFields();
    } catch (error) {
      console.error('ModalCreateTemplate validation error:', error);
    }
  };

  return (
    <Modal
      destroyOnHidden={true}
      styles={{ body: { paddingBottom: '5px' } }}
      centered
      footer={null}
      title={template ? 'Modify template' : 'Create a new template'}
      open={show}
      onCancel={closehandler}
      width="600px"
    >
      <Form
        form={form}
        onFinish={handleFormFinish}
        onSubmitCapture={handleFormSubmit}
        initialValues={getInitialValues(template)}
      >
        <Form.Item
          {...fullLayout}
          name="name"
          className="mt-1"
          required
          validateTrigger="onChange"
          rules={[
            {
              required: true,
              message: 'Please enter template name',
            },
            {
              validator: validateName,
            },
          ]}
        >
          <Input placeholder="Insert template name" allowClear />
        </Form.Item>

        <EnvironmentList
          availableImages={availableImages}
          resources={{
            cpu: cpuInterval,
            ram: ramInterval,
            disk: diskInterval,
          }}
          sharedVolumes={sharedVolumes}
          isPersonal={isPersonal === undefined ? false : isPersonal}
        />

        <div className="flex justify-end gap-2">
          <Button htmlType="submit" onClick={() => closehandler()}>
            Cancel
          </Button>

          <Form.Item shouldUpdate>
            {() => {
              const fieldsError = form.getFieldsError();
              const hasErrors = fieldsError.some(
                ({ errors }) => errors.length > 0,
              );

              return (
                <Button htmlType="submit" type="primary" disabled={hasErrors}>
                  {!loading && (template ? 'Modify' : 'Create')}
                </Button>
              );
            }}
          </Form.Item>
        </div>
      </Form>
    </Modal>
  );
};

export type { TemplateForm as Template };
export default ModalCreateTemplate;
