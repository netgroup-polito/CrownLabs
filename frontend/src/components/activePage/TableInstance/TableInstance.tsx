import { DeleteOutlined } from '@ant-design/icons';
import { Button, Grid, Table } from 'antd';
import { type FC, useCallback, useContext, useMemo, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useDeleteInstanceMutation } from '../../../generated-types';
import { TenantContext } from '../../../contexts/TenantContext';
import {
  formatRelativeDate,
  type Instance,
  WorkspaceRole,
} from '../../../utils';
import ModalGroupDeletion from '../ModalGroupDeletion/ModalGroupDeletion';
import RowInstanceActions, {
  formatElapsedTime,
} from './RowInstanceActions/RowInstanceActions';
import RowInstanceHeader from './RowInstanceHeader/RowInstanceHeader';
import RowInstanceTitle from './RowInstanceTitle/RowInstanceTitle';
import './TableInstance.less';
import RowInstanceStatus from './RowInstanceStatus/RowInstanceStatus';
import UtilsButton from './UtilsButton';

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
  workspacePrettyName?: { [k: string]: string };
}

const TableInstance: FC<ITableInstanceProps> = ({
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
  workspacePrettyName,
}) => {
  const { now } = useContext(TenantContext);
  const [showAlert, setShowAlert] = useState(false);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const destroyAll = () => {
    const deletePromises = instances
      .filter(i => i.persistent === false)
      .map(instance =>
        deleteInstanceMutation({
          variables: {
            instanceId: instance.name,
            tenantNamespace: instance.tenantNamespace!,
          },
        }),
      );

    // Wait for all deletions to complete
    Promise.allSettled(deletePromises);
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

  const screensize = Grid.useBreakpoint();

  const formatWorkspaceName = useCallback(
    (workspaceName: string) =>
      workspacePrettyName ? workspacePrettyName[workspaceName] : workspaceName,
    [workspacePrettyName],
  );

  const extraItem = useMemo(() => {
    let hiddenCols = 0;

    if (screensize.xxl || screensize.xl) return [];
    else if (screensize.lg) hiddenCols = 3;
    else if (screensize.sm) hiddenCols = 5;
    else if (screensize.xs) hiddenCols = 6;

    return [
      {
        title: 'Extra',
        width: '5em',
        render: (_: any, instance: Instance) => (
          <UtilsButton
            age={
              hiddenCols > 0 &&
              formatElapsedTime(now, instance.timeStamp, 'unknown')
            }
            lastAccess={
              hiddenCols > 1 && formatRelativeDate(instance.lastActivity, now)
            }
            user={hiddenCols > 2 && instance.tenantDisplayName}
            template={hiddenCols > 3 && instance.templatePrettyName}
            workspace={
              hiddenCols > 4 && formatWorkspaceName(instance.workspaceName)
            }
            instanceName={hiddenCols > 5 && instance.prettyName}
          />
        ),
      },
    ];
  }, [screensize]);

  if (viewMode === WorkspaceRole.manager)
    return (
      <div>
        <Table
          className="rowInstance-bg-color h-10"
          dataSource={instances}
          pagination={false}
          columns={[
            {
              title: 'Status',
              dataIndex: 'status',
              align: 'center',
              render: (status, instance) => (
                <div className="flex justify-center">
                  <RowInstanceStatus
                    status={status}
                    environments={instance.environments}
                  />
                </div>
              ),
              width: '4em',
              className: 'px-0',
            },
            {
              title: 'ID',
              dataIndex: 'tenantId',
              // responsive: ["sm"],
              ellipsis: true,
            },
            {
              title: 'User',
              dataIndex: 'tenantDisplayName',
              responsive: ['xl'],
              ellipsis: true,
            },
            {
              title: 'Instance Name',
              dataIndex: 'prettyName',
              responsive: ['sm'],
              ellipsis: true,
            },
            {
              title: 'Workspace',
              dataIndex: 'workspaceName',
              responsive: ['md'],
              render: workspaceName => formatWorkspaceName(workspaceName),
              // ellipsis: true,
            },
            {
              title: 'Template',
              dataIndex: 'templatePrettyName',
              responsive: ['md'],
              // ellipsis: true,
            },
            ...extraItem,
            {
              title: 'Age',
              dataIndex: 'timeStamp',
              responsive: ['xl'],
              width: '7em',
              render: timeStamp => formatElapsedTime(now, timeStamp, 'unknown'),
            },
            {
              title: 'Last access',
              dataIndex: 'lastActivity',
              responsive: ['xl'],
              width: '9em',
              render: lastActivity => formatRelativeDate(lastActivity, now),
            },
            {
              title: 'Actions',
              dataIndex: '',
              // responsive: ["sm"],
              render: (_, instance) => (
                <RowInstanceActions
                  instance={instance}
                  hasSSHKeys={hasSSHKeys}
                  now={now}
                  fileManager={true}
                  extended={false}
                  viewMode={viewMode}
                />
              ),
              width: screensize.xs ? '180px' : '230px',
              align: 'center',
            },
          ]}
        />
      </div>
    );

  return (
    <>
      <div
        className={`rowInstance-bg-color ${
          viewMode === WorkspaceRole.user && extended
            ? 'cl-table-instance flex-grow flex-col py-0 content-between overflow-auto scrollbar'
            : ''
        }`}
        style={
          viewMode === WorkspaceRole.user && extended
            ? { maxHeight: '50vh' }
            : undefined
        }
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
                  ? 'w-5/6 sm:w-2/3 lg:w-1/2 xl:w-5/12 2xl:w-5/12'
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
                  ? 'w-1/6 sm:w-1/3 lg:w-1/2 xl:w-7/12 2xl:w-7/12'
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
