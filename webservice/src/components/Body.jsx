import React from 'react';
import ProfessorView from './ProfessorView';
import StudentView from './StudentView';

export default function Body(props) {
  const {
    isStudentView,
    registryName,
    retrieveImageList,
    adminGroups,
    templateLabsAdmin,
    instanceLabsAdmin,
    connectAdmin,
    showStatus,
    createNewTemplate,
    start,
    stopAdmin,
    deleteLabTemplate,
    templateLabs,
    instanceLabs,
    connect,
    stop
  } = props;

  return (
    <div
      style={{
        // the height of the container is viewport heigh - header height(70) - footer height(70)
        height: 'calc(100vh - 134px)',
        overflow: 'auto'
      }}
    >
      {isStudentView ? (
        <StudentView
          templateLabs={templateLabs}
          instanceLabs={instanceLabs}
          start={start}
          connect={connect}
          stop={stop}
          showStatus={showStatus}
        />
      ) : (
        <ProfessorView
          registryName={registryName}
          imageList={retrieveImageList}
          adminGroups={adminGroups}
          templateLabs={templateLabsAdmin}
          instanceLabs={instanceLabsAdmin}
          connect={connectAdmin}
          showStatus={showStatus}
          createNewTemplate={createNewTemplate}
          start={start}
          stop={stopAdmin}
          deleteLabTemplate={deleteLabTemplate}
        />
      )}
    </div>
  );
}
