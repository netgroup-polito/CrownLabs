import { DeleteOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import Button from 'antd-button-color';
import { FC, useContext, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useDeleteInstanceMutation } from '../../../generated-types';
import { TenantContext } from '../../../graphql-components/tenantContext/TenantContext';
import { Instance, WorkspaceRole } from '../../../utils';
import { ModalAlert } from '../../common/ModalAlert';
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
  handleSorting?: (sortingType: string, sorting: number) => void;
  handleManagerSorting?: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string
  ) => void;
}

const TableInstance: FC<ITableInstanceProps> = ({ ...props }) => {
  const {
    instances,
    viewMode,
    extended,
    hasSSHKeys,
    showGuiIcon,
    showAdvanced,
    handleSorting,
    handleManagerSorting,
  } = props;

  const { now } = useContext(TenantContext);
  const [showAlert, setShowAlert] = useState(false);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const destroyAll = () => {
    instances.forEach(instance => {
      deleteInstanceMutation({
        variables: {
          instanceId: instance.name,
          tenantNamespace: instance.tenantNamespace!,
        },
      });
    });
  };

  const [{ templateId }] = instances;

  return (
    <>
      <div
        className={`rowInstance-bg-color ${
          viewMode === 'user' && extended
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
          >
            <Column
              title="Header"
              key="header"
              className="p-0"
              render={() => (
                <RowInstanceHeader
                  viewMode={viewMode}
                  handleSorting={handleSorting!}
                  handleManagerSorting={handleManagerSorting!}
                  templateKey={templateId}
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
            viewMode === 'user' && extended ? '' : 'rowInstance-bg-color'
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
                showGuiIcon={showGuiIcon}
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
            type="danger"
            shape="round"
            size="large"
            icon={<DeleteOutlined />}
            onClick={e => {
              e.stopPropagation();
              setShowAlert(true);
            }}
          >
            Destroy All
          </Button>
          <ModalAlert
            headTitle="Destroy All"
            show={showAlert}
            message="Warning"
            description="This operation will delete all your instances and it is not reversible. Do you want to continue?"
            type="warning"
            buttons={[
              <Button
                type="danger"
                shape="round"
                size="middle"
                icon={<DeleteOutlined />}
                className="border-0"
                onClick={() => destroyAll()}
              >
                Destroy All
              </Button>,
            ]}
            setShow={setShowAlert}
          />
        </div>
      )}
    </>
  );
};

export default TableInstance;
