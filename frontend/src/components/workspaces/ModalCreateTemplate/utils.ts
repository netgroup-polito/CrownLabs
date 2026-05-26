import { getEnvVar } from '../../../env';
import { EnvironmentType, type ImagesQuery } from '../../../generated-types';
import type { Template } from './ModalCreateTemplate';
import type { TemplateFormEnv, Image, ImageList, Resources } from './types';


export const internalRegistry = 'harbor.ng.crownlabs.polito.it'; 
export const projectName = [getEnvVar('VITE_APP_CROWNLABS_IMAGELIST_STANDALONE'), getEnvVar('VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS')];

export const formItemLayout = {
  labelcol: { span: 5 },
  wrappercol: { span: 18 },
  style: { marginBottom: 14 },
};

export const getImageNameNoVer = (image: string) => {
  // split on the last ':' to correctly handle registry:port/repo:tag cases
  return image.includes(':') ? image.slice(0, image.lastIndexOf(':')) : image;
};

export const getDefaultTemplate = (resources: Resources): Template => {
  return {
    name: '',
    description: '',
    environments: [getDefaultTemplateEnvironment(resources, 0)],
    deleteAfter: 'never',
    inactivityTimeout: 'never',
    allowPublicExposure: false,
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
    reservedCpu: 50,
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
    .filter(img => {
      const name = img?.metadata?.name;
      if (!name) return false;
      const normalized = name.trim();
      return projectName.some(proj => proj && normalized === proj.trim());
    })
    .map(img => ({
      name: img!.metadata?.name || 'Unnamed List',
      registryName: img!.spec!.registryName,
      images: img!
        .spec!.images.filter(i => i?.name && i?.versions)
        .map(i => ({
          name: i!.name,
          versions: i!.versions.filter(v => v !== null) as string[],
        })),
    }));
};
