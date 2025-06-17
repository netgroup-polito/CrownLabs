import { DeleteOutlined } from '@ant-design/icons';
import { Button, Table } from 'antd';
import { type FC, useContext, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useDeleteInstanceMutation } from '../../../generated-types';
import { TenantContext } from '../../../contexts/TenantContext';
import { type Instance, WorkspaceRole } from '../../../utils';
import ModalGroupDeletion from '../ModalGroupDeletion/ModalGroupDeletion';
import RowInstanceActions from './RowInstanceActions/RowInstanceActions';
import RowInstanceHeader from './RowInstanceHeader/RowInstanceHeader';
import RowInstanceTitle from './RowInstanceTitle/RowInstanceTitle';
import './TableInstance.less';

const { Column } = Table;
export interface ITableInstanceProps {
  viewMode: WorkspaceRole;
  instances: Array<Instance>;
  hasSSHKeys?: boolean;
  showGuiIcon: boolean;
  extended: boolean;
  showAdvanced?: boolean;
  showCheckbox?: boolean;
  handleSorting?: (sortingType: string, sorting: number) => void;
  handleManagerSorting?: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string,
  ) => void;
  selectiveDestroy?: string[];
  selectToDestroy?: (instanceId: string) => void;
}

const TableInstance: FC<ITableInstanceProps> = ({ ...props }) => {
  const {
    instances,
    viewMode,
    extended,
    hasSSHKeys,
    showGuiIcon,
    showAdvanced,
    showCheckbox,
    handleSorting,
    handleManagerSorting,
    selectiveDestroy,
    selectToDestroy,
  } = props;

  const { now } = useContext(TenantContext);
  const [showAlert, setShowAlert] = useState(false);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const destroyAll = () => {
    instances
      .filter(i => i.persistent === false)
      .forEach(instance => {
        deleteInstanceMutation({
          variables: {
            instanceId: instance.name,
            tenantNamespace: instance.tenantNamespace!,
          },
        });
      });
  };

  const disabled = !instances.find(i => i.persistent === false);

  // Filtering from all instances that ones which are included in the "selectiveDestroy" IDs list
  const selectedIn = instances.filter(i => selectiveDestroy?.includes(i.id));

  const checked = !!selectedIn.length;

  const indeterminate = selectedIn.length !== instances.length && checked;

  const selectGroup = () => {
    // Protect from group selection if selectToDestroy is not defined
    if (!selectToDestroy) return;
    // remap each instance to its ID
    const instIds = instances.map(({ id }) => id);
    // Check if some instance in the List is already selected
    // (Remember each TableInstance represents a grouped list of instances that belong to a single Template)
    if (checked)
      instIds.filter(i => indeterminate !== selectiveDestroy?.includes(i));
    instIds.forEach(selectToDestroy);
  };

  const [{ templateId }] = instances;

  return (
    <>
      <div
        className={`rowInstance-bg-color ${
          viewMode === WorkspaceRole.user && extended
            ? 'cl-table-instance flex-grow flex-wrap content-between py-0 overflow-auto scrollbar'
            : ''
        }`}
      >
        {extended && showAdvanced && (
          <Table
            className="rowInstance-bg-color h-10"
            dataSource={[{}]}
            showHeader={false}
            pagination={false}
            rowClassName=""
            rowKey={_i => 1}
          >
            <Column
              title="Header"
              key="header"
              className="p-0"
              render={() => (
                <RowInstanceHeader
                  viewMode={viewMode}
                  handleSorting={handleSorting!}
                  showCheckbox={showCheckbox || false}
                  handleManagerSorting={handleManagerSorting!}
                  templateKey={templateId}
                  checked={checked}
                  selectGroup={selectGroup}
                  indeterminate={indeterminate}
                />
              )}
            />
          </Table>
        )}
        <Table
          className="rowInstance-bg-color"
          dataSource={instances}
          showHeader={false}
          pagination={false}
          size="middle"
          rowClassName={
            viewMode === WorkspaceRole.user && extended
              ? ''
              : 'rowInstance-bg-color'
          }
          rowKey={record => record.id + (record.templateId || '')}
        >
          <Column
            className={
              extended
                ? viewMode === WorkspaceRole.user
                  ? 'w-5/6 sm:w-2/3 lg:w-3/5 xl:w-1/2 2xl:w-5/12'
                  : 'w-1/2 md:w-2/3 lg:w-7/12 xl:w-1/2'
                : 'w-2/3 md:w-3/4'
            }
            title="Instance Title"
            key="title"
            render={(instance: Instance) => (
              <RowInstanceTitle
                viewMode={viewMode}
                extended={extended}
                instance={instance}
                showCheckbox={showCheckbox}
                showGuiIcon={showGuiIcon}
                selectiveDestroy={selectiveDestroy}
                selectToDestroy={selectToDestroy}
              />
            )}
          />
          <Column
            className={
              extended
                ? viewMode === WorkspaceRole.user
                  ? 'w-1/6 sm:w-1/3 lg:w-2/5 xl:w-1/2 2xl:w-7/12'
                  : 'w-1/2 md:w-1/3 lg:w-5/12 xl:w-1/2'
                : 'w-1/3 md:w-1/4'
            }
            title="Instance Actions"
            key="actions"
            render={(instance: Instance) => (
              <RowInstanceActions
                instance={instance}
                hasSSHKeys={hasSSHKeys}
                now={now}
                fileManager={true}
                extended={extended}
                viewMode={viewMode}
              />
            )}
          />
        </Table>
      </div>
      {extended && viewMode === WorkspaceRole.user && (
        <div className="w-full pt-5 flex justify-center ">
          <Button
            color="danger"
            shape="round"
            size="large"
            icon={<DeleteOutlined />}
            onClick={e => {
              e.stopPropagation();
              setShowAlert(true);
            }}
            disabled={disabled}
          >
            Destroy All
          </Button>
          <ModalGroupDeletion
            view={WorkspaceRole.user}
            persistent={!!instances.find(i => i.persistent === true)}
            selective={false}
            instanceList={instances.map(i => i.id)}
            show={showAlert}
            setShow={setShowAlert}
            destroy={destroyAll}
          />
        </div>
      )}
    </>
  );
};

export default TableInstance;
