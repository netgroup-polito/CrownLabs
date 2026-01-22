import React, { useEffect, useState, useRef } from 'react';
import { Spin, Button, Alert } from 'antd';
import { LoadingOutlined, FolderOpenOutlined } from '@ant-design/icons';
import { useMydrive } from '../../../hooks/useMydrive';
import { Phase2 } from '../../../generated-types';
import './DriveView.css';

const DriveView: React.FC = () => {
  const { mydriveInstance, startDriveInstance, getDriveUrl, instancesLoaded } =
    useMydrive();
  const [iframeLoading, setIframeLoading] = useState(true);
  const [iframeError, setIframeError] = useState(false);
  const hasStartedRef = useRef(false);
  const iframeRef = useRef<HTMLIFrameElement>(null);

  const driveUrl = getDriveUrl();

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
    setIframeLoading(false);
    // Try to detect if iframe loaded successfully or was blocked
    try {
      // This will throw if cross-origin or blocked
      const iframeDoc = iframeRef.current?.contentDocument;
      if (iframeDoc === null) {
        // null means cross-origin (potentially blocked)
        setIframeError(true);
      }
    } catch {
      // Cross-origin error - might be working fine
      setIframeError(false);
    }
  };

  const handleIframeError = () => {
    setIframeLoading(false);
    setIframeError(true);
  };

  const openInNewTab = () => {
    if (driveUrl) {
      window.open(driveUrl, '_blank');
    }
  };

  if (
    !mydriveInstance ||
    mydriveInstance.status !== Phase2.Ready ||
    !driveUrl
  ) {
    return (
      <div className="drive-view-container drive-view-loading">
        <Spin
          indicator={<LoadingOutlined style={{ fontSize: 48 }} spin />}
          tip={
            <div style={{ marginTop: 16 }}>
              <p>Starting the drive instance...</p>
              <p style={{ fontSize: 12, color: '#888' }}>
                Drive instance is loading. This may take a few moments.
              </p>
            </div>
          }
        />
        <div style={{ marginTop: 10, color: 'rgba(0,0,0,0.65)' }}>
          Please be patient while the drive loads, it will take a few moments.
          It will open automatically when ready.
        </div>
      </div>
    );
  }

  // If iframe is blocked, show a fallback with button to open in new tab
  if (iframeError) {
    return (
      <div className="drive-view-container drive-view-loading">
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
      </div>
    );
  }

  return (
    <div className="drive-view-container">
      {iframeLoading && (
        <div className="drive-view-loading-overlay">
          <Spin
            indicator={<LoadingOutlined style={{ fontSize: 32 }} spin />}
            tip="Loading drive..."
          />
          <div style={{ marginTop: 10, color: 'rgba(0,0,0,0.65)' }}>
            Please be patient while the drive loads, it will take a few moments.
            It will open automatically when ready.
          </div>
        </div>
      )}
      <iframe
        ref={iframeRef}
        src={driveUrl}
        className="drive-view-iframe"
        onLoad={handleIframeLoad}
        onError={handleIframeError}
        title="CrownLabs Drive"
      />
    </div>
  );
};

export default DriveView;
