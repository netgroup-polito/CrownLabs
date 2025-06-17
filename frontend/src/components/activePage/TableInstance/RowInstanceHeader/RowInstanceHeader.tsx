import { type FC, useCallback, useMemo, useState } from 'react';
import { Checkbox, Typography } from 'antd';
import { WorkspaceRole } from '../../../../utils';
import {
  CaretDownOutlined,
  CaretUpOutlined,
  SortAscendingOutlined,
  SortDescendingOutlined,
} from '@ant-design/icons';

const { Text } = Typography;
export interface IRowInstanceHeaderProps {
  viewMode: WorkspaceRole;
  templateKey: string;
  handleSorting: (sortingType: string, sorting: number) => void;
  handleManagerSorting: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string,
  ) => void;
  showCheckbox: boolean;
  checked: boolean;
  selectGroup: () => void;
  indeterminate: boolean;
}

const RowInstanceHeader: FC<IRowInstanceHeaderProps> = ({ ...props }) => {
  const {
    viewMode,
    handleSorting,
    handleManagerSorting,
    templateKey: tkey,
    showCheckbox,
    checked,
    selectGroup,
    indeterminate,
  } = props;
  const [prettyNameOrder, setPrettyNameOrder] = useState(0);
  const [templatePrettyNameOrder, setTemplatePrettyNameOrder] = useState(0);
  const [timeStampOrder, setTimeStampOrder] = useState(0);
  const [tenantDisplayNameOrder, setTenantDisplayNameOrder] = useState(0);
  const [tenantIdOrder, setTenantIdOrder] = useState(0);

  type sortKey =
    | 'tenantId'
    | 'tenantDisplayName'
    | 'templatePrettyName'
    | 'timeStamp'
    | 'prettyName';

  const varByKey = useMemo(
    () => ({
      tenantId: tenantIdOrder,
      tenantDisplayName: tenantDisplayNameOrder,
      templatePrettyName: templatePrettyNameOrder,
      timeStamp: timeStampOrder,
      prettyName: prettyNameOrder,
    }),
    [
      tenantIdOrder,
      tenantDisplayNameOrder,
      templatePrettyNameOrder,
      timeStampOrder,
      prettyNameOrder,
    ],
  );

  const setSort = useCallback(
    (key: sortKey) => {
      // get a value for sorting direction
      const getSorting = (value: number) => (value <= 0 ? 1 : -1);

      setTemplatePrettyNameOrder(
        key === 'templatePrettyName' ? getSorting(varByKey[key]) : 0,
      );
      setTimeStampOrder(key === 'timeStamp' ? getSorting(varByKey[key]) : 0);

      setTenantIdOrder(key === 'tenantId' ? getSorting(varByKey[key]) : 0);
      setPrettyNameOrder(key === 'prettyName' ? getSorting(varByKey[key]) : 0);
      setTenantDisplayNameOrder(
        key === 'tenantDisplayName' ? getSorting(varByKey[key]) : 0,
      );
    },
    [varByKey],
  );

  const selectOrder = useCallback(
    (key: sortKey) => {
      // Move setSort inside useCallback to avoid dependency issues
      setSort(key);
      if (viewMode === WorkspaceRole.manager) {
        handleManagerSorting(key, varByKey[key], tkey);
      } else {
        handleSorting(key, varByKey[key]);
      }
    },
    [handleManagerSorting, handleSorting, tkey, viewMode, varByKey, setSort],
  );

  const getArrow = (value: number, alpha: boolean) => {
    if (value > 0) {
      if (alpha) return <SortAscendingOutlined className="pl-1" />;
      return <CaretDownOutlined className="pl-1" />;
    }
    if (value < 0) {
      if (alpha) return <SortDescendingOutlined className="pl-1" />;
      return <CaretUpOutlined className="pl-1" />;
    }
  };

  return (
    <div className="w-full flex justify-between items-center h-10 rowHeader-bg-color">
      <div
        className={
          viewMode === WorkspaceRole.user
            ? 'w-5/6 sm:w-2/3 lg:w-3/5 xl:w-1/2 2xl:w-5/12'
            : 'w-2/3 sm:w-1/2 md:w-2/3 lg:w-7/12 xl:w-1/2'
        }
        title="Instance Title"
        key="title"
      >
        <div className="w-full flex justify-start items-center">
          {showCheckbox && (
            <div className={`flex items-center justify-center w-12`}>
              <Checkbox
                checked={checked}
                className="p-0"
                indeterminate={indeterminate && checked}
                onClick={selectGroup}
              />
            </div>
          )}
          <div
            className={`flex items-center justify-center ${
              showCheckbox ? 'w-14' : 'w-16'
            }`}
          >
            <Text strong>Status</Text>
          </div>
          {viewMode === WorkspaceRole.manager ? (
            <div className="flex items-center justify-center">
              <div
                className="flex items-center justify-start sm:w-36 cursor-pointer"
                onClick={() => selectOrder('tenantId')}
              >
                <Text strong>ID</Text>
                {getArrow(tenantIdOrder, true)}
              </div>
              <div
                className="flex items-center justify-start w-36 2xl:w-44 hidden md:block cursor-pointer "
                onClick={() => selectOrder('tenantDisplayName')}
              >
                <Text strong ellipsis>
                  User
                </Text>
                {getArrow(tenantDisplayNameOrder, true)}
              </div>
              <div
                className="flex items-center justify-start hidden lg:block cursor-pointer"
                onClick={() => selectOrder('prettyName')}
              >
                <Text ellipsis strong>
                  Instance Name
                </Text>
                {getArrow(prettyNameOrder, true)}
              </div>
            </div>
          ) : (
            <>
              <div
                className="flex items-center justify-start w-44 lg:w-52 pl-8 cursor-pointer"
                onClick={() => selectOrder('prettyName')}
              >
                <Text strong>Instance Name</Text>
                {getArrow(prettyNameOrder, true)}
              </div>
              <div
                className="flex items-center justify-start md:w-max hidden xs:block xs:w-28 sm:hidden md:block cursor-pointer"
                onClick={() => selectOrder('templatePrettyName')}
              >
                <Text strong>Template Name</Text>
                {getArrow(templatePrettyNameOrder, true)}
              </div>
            </>
          )}
        </div>
      </div>
      <div
        className={
          viewMode === WorkspaceRole.user
            ? 'w-1/6 sm:w-1/3 lg:w-2/5 xl:w-1/2 2xl:w-7/12'
            : 'w-1/3 sm:w-1/2 md:w-1/3 lg:w-5/12 xl:w-1/2'
        }
      >
        <div className="w-full flex items-center justify-end sm:justify-between">
          <div
            className={`flex justify-between items-center ${
              viewMode === WorkspaceRole.manager
                ? 'lg:w-2/5 xl:w-7/12 2xl:w-1/2'
                : 'lg:w-1/3 xl:w-1/2'
            }`}
          >
            <div className="flex items-center justify-center hidden sm:block w-12 xl:w-40 text-center">
              <Text strong>Utils</Text>
            </div>

            <div
              className="flex items-center justify-center w-12 hidden lg:block text-center cursor-pointer"
              onClick={() => selectOrder('timeStamp')}
            >
              <Text strong>Age</Text>
              {getArrow(timeStampOrder, false)}
            </div>
          </div>
          <div
            className={`flex justify-end items-center gap-2 w-full ${
              viewMode === WorkspaceRole.manager
                ? 'lg:w-3/5 xl:w-5/12 2xl:w-1/2'
                : 'lg:w-2/3 xl:w-1/2'
            }`}
          >
            <div className="flex items-center justify-center w-20 sm:w-40 lg:w-56 xl:w-48">
              <Text strong>Actions</Text>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RowInstanceHeader;
