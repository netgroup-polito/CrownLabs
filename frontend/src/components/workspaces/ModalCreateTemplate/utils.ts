import {
  VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS,
  VITE_APP_CROWNLABS_IMAGELIST_STANDALONE,
  VITE_APP_CROWNLABS_IMAGELIST_PUBLIC_SNAPSHOTS,
} from '../../../env';
import { EnvironmentType, type ImagesQuery } from '../../../generated-types';
import type { Template } from './ModalCreateTemplate';
import type { TemplateFormEnv, Image, ImageList, Resources } from './types';
import { useEffect, useState } from 'react';

export const internalRegistry = 'harbor.ng.crownlabs.polito.it';
export const imageListContainderDisksDefault =
  'harbor-containerdisks-pre-production';
export const imageListStandaloneDefault = 'harbor-standalone-pre-production';
export const projectName = [
  VITE_APP_CROWNLABS_IMAGELIST_STANDALONE,
  VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS,
  imageListStandaloneDefault,
  imageListContainderDisksDefault,
];
export const defaultProjectNameVM = 'crownlabs-containerdisks';
export const defaultProjectNameContainer = 'crownlabs-standalone';
export const formItemLayout = {
  labelcol: { span: 5 },
  wrappercol: { span: 18 },
  style: { marginBottom: 14 },
};

export const getImageNameNoVer = (image: string) => {
  // split on the last ':' to correctly handle registry:port/repo:tag cases
  return image.includes(':') ? image.slice(0, image.lastIndexOf(':')) : image;
};

// Snapshot volumes originate from the template Disk field, which stores an
// integer as a Gi quantity (for example, 15 becomes "15Gi").
export const volumeSizeToGiB = (
  volumeSize: string | number | null | undefined,
): number | undefined => {
  if (volumeSize === null || volumeSize === undefined) return;

  if (typeof volumeSize === 'number') {
    return Number.isInteger(volumeSize) ? volumeSize : undefined;
  }

  const match = volumeSize.trim().match(/^(\d+)Gi$/);
  return match ? Number(match[1]) : undefined;
};

export const getSnapshotDateTime = (
  resourceName?: string | null,
  imageName?: string | null,
): string | undefined => {
  if (!resourceName || !imageName) return;

  const prefix = `${imageName}-`;
  const suffix = resourceName.startsWith(prefix)
    ? resourceName.slice(prefix.length)
    : '';

  return /^\d{8}-\d{6}$/.test(suffix) ? suffix : undefined;
};

export const getDefaultTemplate = (resources: Resources): Template => {
  return {
    name: '',
    description: '',
    environments: [getDefaultTemplateEnvironment(resources, 0)],
    cleanup: {
      deleteAfterCreation: 'never',
      stopAfterInactivity: 'never',
      deleteAfterInactivity: 'never',
    },
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
  return data.imageList.images.flatMap(img => {
    const spec = img?.spec;
    const name = img?.metadata?.name;
    if (!name || !spec?.registryName) return [];

    const normalized = name.trim();
    if (!projectName.some(proj => proj && normalized === proj.trim())) {
      return [];
    }

    return [
      {
        name,
        registryName: spec.registryName,
        projectBaseName: spec.projectBaseName || undefined,
        images: spec.images
          .filter((image): image is NonNullable<typeof image> =>
            Boolean(image?.name),
          )
          .map(image => ({
            name: image.name,
            versions: image.versions.filter(
              (version): version is string => version !== null,
            ),
            versionDetails: image.versionDetails
              ?.filter(
                (detail): detail is NonNullable<typeof detail> =>
                  detail !== null,
              )
              .map(detail => ({
                version: detail.version,
                volumeSize: detail.volumeSize ?? undefined,
              })),
          })),
      },
    ];
  });
};

// Public local images are published by the operator as a cluster-scoped
// ImageList. Unlike workspace images, this list is a catalog: its registryName
// contains the namespace of the public PVCs and versions identify each image.
export const getPublicSnapshotImageList = (
  data: ImagesQuery,
): ImageList | undefined => {
  const imageList = data?.imageList?.images?.find(
    image =>
      image?.metadata?.name === VITE_APP_CROWNLABS_IMAGELIST_PUBLIC_SNAPSHOTS,
  );

  if (!imageList?.spec?.registryName || !imageList.spec.images) return;

  return {
    name: imageList.metadata?.name ?? '',
    registryName: imageList.spec.registryName,
    projectBaseName: imageList.spec.projectBaseName || undefined,
    images: imageList.spec.images
      .filter((image): image is NonNullable<typeof image> =>
        Boolean(image?.name),
      )
      .map(image => ({
        name: image.name,
        versions: image.versions.filter(
          (version): version is string => version !== null,
        ),
        versionDetails: image.versionDetails
          ?.filter(
            (detail): detail is NonNullable<typeof detail> => detail !== null,
          )
          .map(detail => ({
            version: detail.version,
            volumeSize: detail.volumeSize ?? undefined,
          })),
      })),
  };
};

export const useImageLists = (dataImages: ImagesQuery) => {
  const [availableImagesVM, setAvailableImagesVM] = useState<Image[]>([]);
  const [availableImagesContainer, setAvailableImagesContainer] = useState<
    Image[]
  >([]);
  const [projectBaseNameVM, setProjectBaseNameVM] = useState<string>('');
  const [projectBaseNameContainer, setProjectBaseNameContainer] =
    useState<string>('');

  useEffect(() => {
    if (!dataImages) {
      setAvailableImagesVM([]);
      setAvailableImagesContainer([]);
      return;
    }

    const imageLists = getImageLists(dataImages);
    const internalImagesVM =
      imageLists.find(
        list => list.name === VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS,
      ) ||
      imageLists.find(list => list.name === imageListContainderDisksDefault);
    setProjectBaseNameVM(
      internalImagesVM?.projectBaseName || defaultProjectNameVM,
    );

    const internalImagesContainer =
      imageLists.find(
        list => list.name === VITE_APP_CROWNLABS_IMAGELIST_STANDALONE,
      ) || imageLists.find(list => list.name === imageListStandaloneDefault);

    setProjectBaseNameContainer(
      internalImagesContainer?.projectBaseName || defaultProjectNameContainer,
    );

    if (!internalImagesVM) {
      setAvailableImagesVM([]);
      return;
    }

    if (!internalImagesContainer) {
      setAvailableImagesContainer([]);
      return;
    }
    setAvailableImagesContainer(getImagesFromList(internalImagesContainer));
    setAvailableImagesVM(getImagesFromList(internalImagesVM));
  }, [dataImages]);

  return {
    availableImagesVM,
    availableImagesContainer,
    projectBaseNameVM,
    projectBaseNameContainer,
  };
};
export const isInImageList = (
  image: string,
  envType: string,
  availableImagesVM: Image[],
  availableImagesContainer: Image[],
): boolean => {
  const parsedImge = getImageNameNoVer(image).split('/').slice(-1).join('/');
  if (envType === EnvironmentType.VirtualMachine) {
    return availableImagesVM.some(img => {
      const imgNameNoVer = getImageNameNoVer(img.name)
        .split('/')
        .slice(-1)
        .join('/');
      return imgNameNoVer === parsedImge;
    });
  } else if (envType === EnvironmentType.Standalone) {
    return availableImagesContainer.some(img => {
      const imgNameNoVer = getImageNameNoVer(img.name)
        .split('/')
        .slice(-1)
        .join('/');
      return imgNameNoVer === parsedImge;
    });
  }
  return false;
};
