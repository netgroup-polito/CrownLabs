import { EnvironmentType, type ImagesQuery } from '../../../generated-types';
import type { Template } from './ModalCreateTemplate';
import type { TemplateFormEnv, Image, ImageList, Resources } from './types';

export const internalRegistry = 'registry.internal.crownlabs.polito.it';

export const formItemLayout = {
  labelCol: { span: 5 },
  wrapperCol: { span: 19 },
};

export const getImageNameNoVer = (image: string) => {
  // split on the last ':' to correctly handle registry:port/repo:tag cases
  return image.includes(':') ? image.slice(0, image.lastIndexOf(':')) : image;
};

export const getDefaultTemplate = (resources: Resources): Template => {
  return {
    name: '',
    environments: [getDefaultTemplateEnvironment(resources, 0)],
    deleteAfter: 'never',
    inactivityTimeout: 'never',
  };
};

export const getDefaultTemplateEnvironment = (
  resources: Resources,
  envIndex: number,
): TemplateFormEnv => {
  return {
    name: `env-${envIndex + 1}`,
    image: '',
    registry: '',
    environmentType: EnvironmentType.VirtualMachine,
    persistent: false,
    gui: true,
    cpu: resources.cpu.min,
    ram: resources.ram.min,
    disk: 0,
    sharedVolumeMounts: [],
    rewriteUrl: false,
  };
};

// Get images from selected image list
export const getImagesFromList = (imageList: ImageList): Image[] => {
  const images: Image[] = [];

  imageList.images.forEach(img => {
    const versionsInImageName: Image[] = img.versions.map(v => ({
      name: `${img.name}:${v}`,
      type: [],
      registry: imageList.registryName,
    }));

    images.push(...versionsInImageName);
  });

  return images;
};

// Process image lists from the query
export const getImageLists = (data: ImagesQuery): ImageList[] => {
  if (!data?.imageList?.images) return [];

  return data.imageList.images
    .filter(img => img?.spec?.registryName && img?.spec?.images)
    .map(img => ({
      name: img!.spec!.registryName,
      registryName: img!.spec!.registryName,
      images: img!
        .spec!.images.filter(i => i?.name && i?.versions)
        .map(i => ({
          name: i!.name,
          versions: i!.versions.filter(v => v !== null) as string[],
        })),
    }));
};
