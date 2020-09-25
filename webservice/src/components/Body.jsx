import React from 'react';
import ProfessorView from '../views/ProfessorView';
import StudentView from '../views/StudentView';

export default function Body(props) {
  const {
    isStudentView,
    registryName,
    retriveImageList,
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
          imageList={retriveImageList}
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
