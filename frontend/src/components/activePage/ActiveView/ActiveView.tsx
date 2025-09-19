import { Col, Space } from 'antd';
import type { FC } from 'react';
import { useEffect, useState } from 'react';
import type { User, Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { SessionValue, StorageKeys } from '../../../utilsStorage';
import Box from '../../common/Box';
import ModalGroupDeletion from '../ModalGroupDeletion/ModalGroupDeletion';
import TableInstanceLogic from '../TableInstance/TableInstanceLogic';
import TableWorkspaceLogic from '../TableWorkspaceLogic/TableWorkspaceLogic';
import Toolbox from '../Toolbox/Toolbox';
import ViewModeButton from './ViewModeButton/ViewModeButton';
import QuotaDisplay from '../../workspaces/QuotaDisplay/QuotaDisplay'; // Import QuotaDisplay
import { QuotaProvider } from '../../../contexts/QuotaContext';
const view = new SessionValue(StorageKeys.Active_View, WorkspaceRole.user);
const advanced = new SessionValue(StorageKeys.Active_Headers, 'true');

export interface IActiveViewProps {
  user: User;
  workspaces: Array<Workspace>;
  managerView: boolean;
  quotaData: {
    consumedQuota: {
      cpu?: string | number;
      memory?: string;
      instances?: number;
    };
    workspaceQuota: {
      cpu?: string | number;
      memory?: string;
      instances?: number;
    };
    availableQuota: {
      cpu?: string | number;
      memory?: string;
      instances?: number;
    };
    showQuotaDisplay: boolean;
    refreshQuota?: () => void; // Add refresh function
  };
}

const ActiveView: FC<IActiveViewProps> = ({ ...props }) => {
  const { managerView, user, workspaces, quotaData } = props;

  const [expandAll, setExpandAll] = useState(false);
  const [collapseAll, setCollapseAll] = useState(false);
  const [destroySelectedTrigger, setDestroySelectedTrigger] = useState(false);
  const [showAlert, setShowAlert] = useState(false);
  const [searchField, setSearchField] = useState('');
  const [currentView, setCurrentView] = useState<WorkspaceRole>(
    managerView ? (view.get() as WorkspaceRole) : WorkspaceRole.user,
  );
  const [showAdvanced, setShowAdvanced] = useState(
    !managerView || advanced.get() !== 'false',
  );
  const [showCheckbox, setShowCheckbox] = useState(false);
  const [selectiveDestroy, setSelectiveDestroy] = useState<string[]>([]);
  const [selectedPersistent, setSelectedPersistent] = useState<boolean>(false);

  const selectToDestroy = (instanceId: string) =>
    selectiveDestroy.includes(instanceId)
      ? setSelectiveDestroy(old => old.filter(id => id !== instanceId))
      : setSelectiveDestroy(old => [...old, instanceId]);

  const deselectAll = () => setSelectiveDestroy([]);

  const displayCheckbox = () => {
    if (!showCheckbox) {
      setShowCheckbox(true);
    } else {
      setShowCheckbox(() => {
        deselectAll();
        return false;
      });
    }
  };

  useEffect(() => {
    view.set(currentView);
  }, [currentView]);

  useEffect(() => {
    advanced.set(String(showAdvanced));
  }, [showAdvanced]);

  return (
    <QuotaProvider
      refreshQuota={quotaData?.refreshQuota}
      availableQuota={quotaData?.availableQuota}
    >
      <Col span={24} style={{ height: '85vh' }}>
        <ModalGroupDeletion
          view={WorkspaceRole.manager}
          persistent={selectedPersistent}
          selective={true}
          instanceList={selectiveDestroy}
          show={showAlert}
          setShow={setShowAlert}
          destroy={() => setDestroySelectedTrigger(true)}
        />

        {/* Remove the wrapper div with fixed height and use full height */}
        <Box
          header={{
            center: !managerView ? (
              <div className="h-full flex justify-center items-center px-5">
                <p className="md:text-2xl text-lg text-center mb-0">
                  <b>Active Instances</b>
                </p>
              </div>
            ) : (
              ''
            ),
            size: 'middle',
            right: managerView && (
              <div className="h-full flex justify-center items-center pr-10">
                <Space size="small">
                  <ViewModeButton
                    setCurrentView={setCurrentView}
                    currentView={currentView}
                  />
                </Space>
              </div>
            ),
            left: managerView && currentView === WorkspaceRole.manager && (
              <div className="h-full flex justify-center items-center pl-6 gap-4">
                <Toolbox
                  setSearchField={setSearchField}
                  setExpandAll={setExpandAll}
                  setCollapseAll={setCollapseAll}
                  showAdvanced={showAdvanced}
                  setShowAdvanced={setShowAdvanced}
                  showCheckbox={showCheckbox}
                  setShowCheckbox={displayCheckbox}
                  setShowAlert={setShowAlert}
                  selectiveDestroy={selectiveDestroy}
                  deselectAll={deselectAll}
                />
              </div>
            ),
          }}
          style={{ height: '100%' }} // Make Box take full height
        >
          {/* Add padding wrapper */}
          <div
            style={{
              padding: '24px',
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            {/* Fixed QuotaDisplay */}
            <div style={{ flexShrink: 0, marginBottom: '16px' }}>
              <QuotaDisplay
                consumedQuota={quotaData.consumedQuota}
                workspaceQuota={quotaData.workspaceQuota}
              />
            </div>

            {/* Scrollable content */}
            <div style={{ flex: 1, overflow: 'auto', minHeight: 0 }}>
              {currentView === WorkspaceRole.manager && managerView ? (
                <TableWorkspaceLogic
                  workspaces={workspaces}
                  user={user}
                  filter={searchField}
                  collapseAll={collapseAll}
                  expandAll={expandAll}
                  setCollapseAll={setCollapseAll}
                  setExpandAll={setExpandAll}
                  showAdvanced={showAdvanced}
                  showCheckbox={showCheckbox}
                  destroySelectedTrigger={destroySelectedTrigger}
                  setDestroySelectedTrigger={setDestroySelectedTrigger}
                  selectiveDestroy={selectiveDestroy}
                  selectToDestroy={selectToDestroy}
                  setSelectedPersistent={setSelectedPersistent}
                />
              ) : (
                <TableInstanceLogic
                  showGuiIcon={true}
                  user={user}
                  viewMode={currentView}
                  extended={true}
                />
              )}
            </div>
          </div>
        </Box>
      </Col>
    </QuotaProvider>
  );
};

export default ActiveView;
