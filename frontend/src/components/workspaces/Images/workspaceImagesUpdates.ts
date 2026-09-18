import {
  UpdateType,
  type UpdatedWorkspaceImagesSubscription,
  type WorkspaceImagesQuery,
} from '../../../generated-types';

export const updateWorkspaceImages = (
  previous: WorkspaceImagesQuery,
  update: UpdatedWorkspaceImagesSubscription['updatedImage'],
  workspaceNamespace: string,
): WorkspaceImagesQuery => {
  const image = update?.image;
  const metadata = image?.metadata;
  if (
    !previous.imageList ||
    !image ||
    !metadata?.name ||
    metadata.namespace !== workspaceNamespace
  ) {
    return previous;
  }

  const images = previous.imageList.images;
  const matches = (item: (typeof images)[number]) =>
    item?.metadata?.name === metadata.name &&
    item?.metadata?.namespace === metadata.namespace;

  if (update?.updateType === UpdateType.Deleted) {
    return {
      ...previous,
      imageList: {
        ...previous.imageList,
        images: images.filter(
          item =>
            !matches(item) ||
            // A delayed delete must not remove a recreated resource.
            Boolean(
              metadata.uid &&
                item?.metadata?.uid &&
                metadata.uid !== item.metadata.uid,
            ),
        ),
      },
    };
  }

  if (
    update?.updateType !== UpdateType.Added &&
    update?.updateType !== UpdateType.Modified
  ) {
    return previous;
  }

  // Replace the matching image, or append it when it is not in the list yet.
  // This also prevents repeated ADDED events from creating duplicate rows.
  return {
    ...previous,
    imageList: {
      ...previous.imageList,
      images: images.some(matches)
        ? images.map(item => (matches(item) ? image : item))
        : [...images, image],
    },
  };
};
