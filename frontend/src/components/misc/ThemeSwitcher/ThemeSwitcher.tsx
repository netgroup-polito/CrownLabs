import { useContext } from 'react';
import { Switch } from 'antd';
import { BulbTwoTone, BulbOutlined } from '@ant-design/icons';
import { ThemeContext } from '../../../contexts/ThemeContext';

// this component doesn't include a story
// since it conflicts with the storybook theme management

const ThemeSwitcher = () => {
  const { isDarkTheme, setIsDarkTheme } = useContext(ThemeContext);

  const onChange = () => {
    setIsDarkTheme(!isDarkTheme);
  };

  return (
    <Switch
      onChange={onChange}
      checked={isDarkTheme}
      checkedChildren={<BulbOutlined />}
      unCheckedChildren={<BulbTwoTone />}
    />
  );
};

export default ThemeSwitcher;
