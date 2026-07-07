import { VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS, VITE_APP_CROWNLABS_IMAGELIST_STANDALONE } from '../../../env';
import { EnvironmentType, type ImagesQuery } from '../../../generated-types';
import type { Template } from './ModalCreateTemplate';
import type { TemplateFormEnv, Image, ImageList, Resources } from './types';
import { useEffect, useState } from 'react';


export const internalRegistry = 'harbor.ng.crownlabs.polito.it'; 
export const imageListContainderDisksDefault = "harbor-containerdisks-pre-production";
export const imageListStandaloneDefault = "harbor-standalone-pre-production";
export const projectName = [VITE_APP_CROWNLABS_IMAGELIST_STANDALONE, VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS, imageListStandaloneDefault, imageListContainderDisksDefault];
export const defaultProjectNameVM = "crownlabs-containerdisks";
export const defaultProjectNameContainer = "crownlabs-standalone";
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
      projectBaseName: img!.spec!.projectBaseName || undefined,
      images: img!
        .spec!.images.filter(i => i?.name && i?.versions)
        .map(i => ({
          name: i!.name,
          versions: i!.versions.filter(v => v !== null) as string[],
        })),
    }));
};


export const useImageLists = (dataImages: ImagesQuery) => {
  const [availableImagesVM, setAvailableImagesVM] = useState<Image[]>([]);
    const [availableImagesContainer, setAvailableImagesContainer] = useState<Image[]>([]);
    const [projectBaseNameVM, setProjectBaseNameVM] = useState<string>("");
    const [projectBaseNameContainer, setProjectBaseNameContainer] = useState<string>("");
  
      useEffect(() => {
          if (!dataImages) {
            setAvailableImagesVM([]);
            setAvailableImagesContainer([]);
            return;
          }
      
          const imageLists = getImageLists(dataImages);
          const internalImagesVM = imageLists.find(
            list => list.name === VITE_APP_CROWNLABS_IMAGELIST_CONTAINERDISKS
          ) || imageLists.find(
            list => list.name === imageListContainderDisksDefault
          );
          setProjectBaseNameVM(internalImagesVM?.projectBaseName || defaultProjectNameVM); 
      
          const internalImagesContainer = imageLists.find(
            list => list.name === VITE_APP_CROWNLABS_IMAGELIST_STANDALONE
          ) || imageLists.find(
            list => list.name === imageListStandaloneDefault
          );
      
          setProjectBaseNameContainer(internalImagesContainer?.projectBaseName || defaultProjectNameContainer);
            
      
      
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

  return { availableImagesVM, availableImagesContainer, projectBaseNameVM, projectBaseNameContainer };
}
export const isInImageList = (image: string, envType: string, availableImagesVM: Image[], availableImagesContainer: Image[]): boolean => {
       const parsedImge = getImageNameNoVer(image).split('/').slice(-1).join('/');
      if(envType === EnvironmentType.VirtualMachine) {
        return availableImagesVM.some(img => {
          const imgNameNoVer = getImageNameNoVer(img.name).split('/').slice(-1).join('/');
          return imgNameNoVer === parsedImge;
        });
      } else if (envType === EnvironmentType.Standalone) {
      return availableImagesContainer.some(img => {
        const imgNameNoVer = getImageNameNoVer(img.name).split('/').slice(-1).join('/');
        return imgNameNoVer === parsedImge;
      });
    };
    return false;
  }