import {
  VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS,
  VITE_APP_CROWNLABS_IMAGELIST_STANDALONE,
  VITE_APP_CROWNLABS_IMAGELIST_PUBLIC_SNAPSHOTS,
} from '../../../env';
import {
  EnvironmentType,
  useImageListQuery,
  type ImageListQuery,
  type ImageListQueryVariables,
} from '../../../generated-types';
import type { Template } from './ModalCreateTemplate';
import type { TemplateFormEnv, Image, ImageList, Resources } from './types';
import { useMemo } from 'react';
import type { ApolloError, QueryHookOptions } from '@apollo/client';

export const internalRegistry = 'harbor.ng.crownlabs.polito.it';
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

export type ImageListQueryOptions = Pick<
  QueryHookOptions<ImageListQuery, ImageListQueryVariables>,
  'skip' | 'fetchPolicy' | 'onError'
>;

// Convert the ImageList returned by the query into the format used by the form
const toImageList = (
  imageList: ImageListQuery['imageList'],
): ImageList | undefined => {
  const name = imageList?.metadata?.name;
  const spec = imageList?.spec;
  if (!name || !spec?.registryName) return;

  return {
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
            (detail): detail is NonNullable<typeof detail> => detail !== null,
          )
          .map(detail => ({
            version: detail.version,
            volumeSize: detail.volumeSize ?? undefined,
          })),
      })),
  };
};

const isNotFoundError = (error: ApolloError) =>
  error.graphQLErrors.some(err => err.extensions?.statusCode === 404);

// Fetch a single ImageList by name. When the name is not set or the ImageList
// does not exist, the list is empty; other errors are forwarded to onError.
const useImageList = (name: string, options: ImageListQueryOptions) => {
  const { data, loading, error } = useImageListQuery({
    variables: { name },
    skip: !name || options.skip,
    fetchPolicy: options.fetchPolicy,
    onError: err => {
      if (isNotFoundError(err)) {
        console.warn(
          `WARNING: IMAGELIST ${name} NOT FOUND, the related image list will be empty`,
        );
        return;
      }
      options.onError?.(err);
    },
  });

  const imageList = useMemo(() => toImageList(data?.imageList), [data]);

  return {
    imageList,
    loading,
    error: error && !isNotFoundError(error) ? error : undefined,
  };
};

// Fetch the ImageLists of VM (containerdisks) and container (standalone)
// images, whose names are set through the environment
export const useImageLists = (options: ImageListQueryOptions = {}) => {
  const { imageList: imageListVM } = useImageList(
    VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS,
    options,
  );
  const { imageList: imageListContainer } = useImageList(
    VITE_APP_CROWNLABS_IMAGELIST_STANDALONE,
    options,
  );

  const availableImagesVM = useMemo(
    () => (imageListVM ? getImagesFromList(imageListVM) : []),
    [imageListVM],
  );
  const availableImagesContainer = useMemo(
    () => (imageListContainer ? getImagesFromList(imageListContainer) : []),
    [imageListContainer],
  );

  return {
    availableImagesVM,
    availableImagesContainer,
    projectBaseNameVM: imageListVM?.projectBaseName || defaultProjectNameVM,
    projectBaseNameContainer:
      imageListContainer?.projectBaseName || defaultProjectNameContainer,
  };
};

// Public local images are published by the operator as a cluster-scoped
// ImageList. Unlike workspace images, this list is a catalog: its registryName
// contains the namespace of the public PVCs and versions identify each image.
export const usePublicSnapshotImageList = (
  options: ImageListQueryOptions = {},
) => {
  const { imageList, loading, error } = useImageList(
    VITE_APP_CROWNLABS_IMAGELIST_PUBLIC_SNAPSHOTS,
    options,
  );

  return {
    publicSnapshotImageList: imageList,
    loadingPublicSnapshotImageList: loading,
    publicSnapshotImageListError: Boolean(error),
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
