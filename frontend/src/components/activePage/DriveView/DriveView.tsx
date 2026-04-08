import React, { useEffect, useState, useRef } from 'react';
import { Spin, Button, Alert } from 'antd';
import { LoadingOutlined, FolderOpenOutlined } from '@ant-design/icons';
import { useMydrive } from './useMydrive';
import { Phase2 } from '../../../generated-types';
import './DriveView.css';

const DriveView: React.FC = () => {
  const { mydriveInstance, startDriveInstance, getDriveUrl, instancesLoaded } =
    useMydrive();
  const [iframeError, setIframeError] = useState(false);
  const hasStartedRef = useRef(false);
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const [readyDelayed, setReadyDelayed] = useState(false);
  const initialCheckDoneRef = useRef(false);

  const driveUrl = getDriveUrl();

  // Handle delay logic:
  // 1. If instance is Ready on first load -> No delay
  // 2. If instance becomes Ready after starting -> 2s delay for ingress
  useEffect(() => {
    if (!instancesLoaded) return;

    // Status can be undefined if no instance exists yet
    const currentStatus = mydriveInstance?.status;

    if (currentStatus === Phase2.Ready) {
      if (!initialCheckDoneRef.current) {
        // Was ready immediately on load - no delay needed
        setReadyDelayed(true);
      } else if (!readyDelayed) {
        // Became ready dynamically - wait for ingress
        const timer = setTimeout(() => {
          setReadyDelayed(true);
        }, 2000);
        // TODO (see issue #1040): after this issue will be solved the timeout can be removed
        return () => clearTimeout(timer);
      }
    } else {
      // Not ready (or failed/off)
      setReadyDelayed(false);
    }

    // Mark that we have performed the initial check
    initialCheckDoneRef.current = true;
  }, [mydriveInstance?.status, instancesLoaded, readyDelayed]);

  // Automatically start the drive if it's not running
  // Use a ref to ensure we only attempt to start once per mount
  // Wait until instances are loaded AND mydriveInstance state is populated
  useEffect(() => {
    // Only proceed if instances are loaded from context
    if (!instancesLoaded) {
      return;
    }

    // Only try to start once
    if (hasStartedRef.current) {
      return;
    }

    // If no instance exists, create one
    if (!mydriveInstance) {
      hasStartedRef.current = true;
      startDriveInstance();
    }
    // If instance is off, start it
    else if (mydriveInstance.status === Phase2.Off) {
      hasStartedRef.current = true;
      startDriveInstance();
    }
    // If instance exists and is ready or starting, don't do anything
    else if (
      mydriveInstance.status === Phase2.Ready ||
      mydriveInstance.status === Phase2.Starting ||
      mydriveInstance.status === Phase2.Importing
    ) {
      hasStartedRef.current = true; // Mark as handled
    }
  }, [mydriveInstance, startDriveInstance, instancesLoaded]);

  const handleIframeLoad = () => {
    setIframeError(false);
  };

  const handleIframeError = () => {
    setIframeError(true);
  };

  const openInNewTab = () => {
    if (driveUrl) {
      window.open(driveUrl, '_blank');
    }
  };

  const isInstanceReady =
    mydriveInstance && mydriveInstance.status === Phase2.Ready;

  let content: React.ReactNode;
  let containerClassName = 'drive-view-container';

  if (!isInstanceReady || !driveUrl || !readyDelayed) {
    containerClassName += ' drive-view-loading';
    content = (
      <>
        <Spin
          indicator={<LoadingOutlined style={{ fontSize: 48 }} spin />}
          tip={
            <div style={{ marginTop: 16 }}>
              <p>
                {isInstanceReady
                  ? 'Connecting to drive...'
                  : 'Starting the drive instance...'}
              </p>
              <p style={{ fontSize: 12, color: '#888' }}>
                {isInstanceReady
                  ? 'Establishing secure connection...'
                  : 'Drive instance is loading. This may take a few moments.'}
              </p>
            </div>
          }
        />
        <div style={{ marginTop: 10 }}>
          <p>
            Please be patient while the drive loads, it will take a few moments.
            It will open automatically when ready.
          </p>
        </div>
      </>
    );
  } else if (iframeError) {
    // If iframe is blocked, show a fallback with button to open in new tab
    containerClassName += ' drive-view-loading';
    content = (
      <>
        <Alert
          message="Sorry, we cannot load the drive here"
          description="You can open it in a new tab using the button below"
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />
        <Button
          type="primary"
          size="large"
          icon={<FolderOpenOutlined />}
          onClick={openInNewTab}
        >
          Open drive in new tab
        </Button>
      </>
    );
  } else {
    containerClassName += ' ant-card ant-layout-content';
    content = (
      <iframe
        ref={iframeRef}
        src={driveUrl}
        className="drive-view-iframe"
        onLoad={handleIframeLoad}
        onError={handleIframeError}
        title="CrownLabs Drive"
      />
    );
  }

  return <div className={containerClassName}>{content}</div>;
};

export default DriveView;
