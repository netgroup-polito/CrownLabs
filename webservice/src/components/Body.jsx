import React from 'react';
import ProfessorView from './ProfessorView';
import StudentView from './StudentView';

export default function Body(props) {
  const {
    isStudentView,
    retrieveImageList,
    adminGroups,
    templateLabsAdmin,
    instanceLabsAdmin,
    connectAdmin,
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
        />
      ) : (
        <ProfessorView
          imageList={retrieveImageList}
          adminGroups={adminGroups}
          templateLabs={templateLabsAdmin}
          instanceLabs={instanceLabsAdmin}
          connect={connectAdmin}
          createNewTemplate={createNewTemplate}
          start={start}
          stop={stopAdmin}
          deleteLabTemplate={deleteLabTemplate}
        />
      )}
    </div>
  );
}
