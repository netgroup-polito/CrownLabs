import React, { Fragment } from 'react';

const ListItemFields = props => {
  const { fields } = props;
  return (
    <>
      {Object.keys(fields).map(fieldName => (
        <Fragment key={fieldName}>
          {fields[fieldName] && (
            <>
              <b>{fieldName}</b>
              <b>: </b>
              {fields[fieldName]}
              <br />
            </>
          )}
        </Fragment>
      ))}
    </>
  );
};

export default ListItemFields;
