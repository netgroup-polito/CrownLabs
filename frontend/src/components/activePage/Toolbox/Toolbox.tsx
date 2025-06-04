import {
  CloseOutlined,
  ControlOutlined,
  DeleteOutlined,
  FullscreenExitOutlined,
  FullscreenOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import { Divider, Input, Popover, Switch } from 'antd';
import { Button } from 'antd';
import { useState, type Dispatch, type FC, type SetStateAction } from 'react';

const { Search } = Input;

export interface IToolboxProps {
  setSearchField: Dispatch<SetStateAction<string>>;
  setExpandAll: Dispatch<SetStateAction<boolean>>;
  setCollapseAll: Dispatch<SetStateAction<boolean>>;
  showAdvanced: boolean;
  setShowAdvanced: Dispatch<SetStateAction<boolean>>;
  showCheckbox: boolean;
  setShowCheckbox: Dispatch<SetStateAction<boolean>>;
  setShowAlert: Dispatch<SetStateAction<boolean>>;
  selectiveDestroy: Array<string>;
  deselectAll: () => void;
}

const Toolbox: FC<IToolboxProps> = ({ ...props }) => {
  const {
    setSearchField,
    setExpandAll,
    setCollapseAll,
    showAdvanced,
    setShowAdvanced,
    showCheckbox,
    setShowCheckbox,
    setShowAlert,
    selectiveDestroy,
    deselectAll,
  } = props;
  const [toolboxPopover, setToolboxPopover] = useState(false);
  const [searchPopover, setSearchPopover] = useState(false);

  const mobileContent = (
    <div className="flex flex-col justify-center gap-2">
      <div className="flex justify-between gap-2 md:hidden">
        <Button
          type="primary"
          shape="round"
          size="middle"
          icon={<FullscreenOutlined />}
          onClick={() => setExpandAll(true)}
        >
          Expand
        </Button>
        <Button
          type="primary"
          shape="round"
          size="middle"
          icon={<FullscreenExitOutlined />}
          onClick={() => setCollapseAll(true)}
        >
          Collapse
        </Button>
      </div>
      <Divider type="horizontal" className="my-2 md:hidden" />
      <div className="flex flex-col justify-start gap-2 xl:hidden">
        <div className="flex items-center gap-2">
          <Switch
            checked={showAdvanced}
            onClick={setShowAdvanced}
            size="small"
          />
          <span>Show Header</span>
        </div>
        <div className="flex items-center gap-2">
          <Switch
            checked={showCheckbox}
            onChange={setShowCheckbox}
            size="small"
          />
          <span>Show Checkbox</span>
        </div>
      </div>

      <Divider type="horizontal" className="my-2 xl:hidden" />
      <Button
        type="primary"
        ghost
        shape="round"
        size="middle"
        icon={<DeleteOutlined />}
        onClick={e => {
          e.stopPropagation();
          setToolboxPopover(false);
          setShowAlert(true);
        }}
        disabled={!selectiveDestroy.length}
      >
        Destroy Selected{` (${selectiveDestroy.length})`}
      </Button>
      <Button
        color={!selectiveDestroy.length ? 'primary' : 'orange'}
        ghost
        shape="round"
        size="middle"
        onClick={() => deselectAll()}
        disabled={!selectiveDestroy.length}
      >
        Deselect All
      </Button>
    </div>
  );
  return (
    <div className="h-full flex justify-start items-center gap-2 px-5 py-3">
      <Search
        allowClear
        className="hidden sm:block"
        placeholder="Search User"
        onChange={event => setSearchField(event.target.value)}
        onSearch={setSearchField}
        enterButton
      />
      <Popover
        className="block sm:hidden"
        content={
          <div className="flex justify-center gap-4">
            <Search
              placeholder="Search User"
              onChange={event => setSearchField(event.target.value)}
              onSearch={setSearchField}
              allowClear
              enterButton
            />
          </div>
        }
        placement="bottomRight"
        trigger="click"
        open={searchPopover}
        onOpenChange={() => setSearchPopover(old => !old)}
      >
        <Button
          color={searchPopover ? 'danger' : 'primary'}
          shape="circle"
          size="middle"
          icon={!searchPopover ? <SearchOutlined /> : <CloseOutlined />}
        />
      </Popover>
      <Divider type="vertical" className="hidden md:block" />
      <Button
        className="hidden md:block"
        type="primary"
        shape="round"
        size="middle"
        icon={<FullscreenOutlined />}
        onClick={() => setExpandAll(true)}
      >
        Expand
      </Button>
      <Button
        className="hidden md:block"
        type="primary"
        shape="round"
        size="middle"
        icon={<FullscreenExitOutlined />}
        onClick={() => setCollapseAll(true)}
      >
        Collapse
      </Button>
      <Divider type="vertical" className="hidden xl:block" />
      <div className="flex flex-col justify-start gap-2 hidden xl:block">
        <div className="flex items-center w-36 gap-2">
          <Switch
            checked={showAdvanced}
            onClick={setShowAdvanced}
            size="small"
          />
          <span>Show Header</span>
        </div>
        <div className="flex items-center w-36 gap-2">
          <Switch
            checked={showCheckbox}
            onChange={setShowCheckbox}
            size="small"
          />
          <span>Show Checkbox</span>
        </div>
      </div>

      <Divider type="vertical" className="" />
      <Button
        className="hidden 2xl:block"
        type="primary"
        ghost
        shape="round"
        size="middle"
        icon={<DeleteOutlined />}
        onClick={e => {
          e.stopPropagation();
          setShowAlert(true);
        }}
        disabled={!selectiveDestroy.length}
      >
        Destroy Selected{` (${selectiveDestroy.length})`}
      </Button>
      <Button
        className="hidden 2xl:block"
        color={!selectiveDestroy.length ? 'primary' : 'orange'}
        ghost
        shape="round"
        size="middle"
        onClick={() => deselectAll()}
        disabled={!selectiveDestroy.length}
      >
        Deselect All
      </Button>
      <Popover
        className="block 2xl:hidden"
        content={mobileContent}
        placement="bottom"
        trigger="click"
        open={toolboxPopover}
        onOpenChange={() => setToolboxPopover(old => !old)}
      >
        <Button
          color={toolboxPopover ? 'danger' : 'primary'}
          ghost
          shape="circle"
          size="middle"
          icon={!toolboxPopover ? <ControlOutlined /> : <CloseOutlined />}
        />
      </Popover>
    </div>
  );
};

export default Toolbox;
