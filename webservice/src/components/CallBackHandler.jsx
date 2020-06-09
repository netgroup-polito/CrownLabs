import React from 'react';
/**
 * Dumb component to manage the callback. It renders nothing, just call the function passed as props
 * @param props the function to be executed
 */
export default function CallBackHandler(props) {
  const { func } = props;
  func();
  return <div />;
}
