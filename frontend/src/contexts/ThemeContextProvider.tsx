import { StyleProvider } from '@ant-design/cssinjs';
import { ConfigProvider, theme } from 'antd';
import { type FC, type PropsWithChildren, useEffect, useState } from 'react';
import { ThemeContext } from './ThemeContext';

const LOCALSTORAGE_IS_DARK_THEME_KEY = 'VITE_APP_IS_DARK_THEME';
const MEDIA_QUERY_PREFER_LIGHT_SCHEMA = '(prefers-color-scheme: light)';
const MEDIA_QUERY_PREFER_DARK_SCHEMA = '(prefers-color-scheme: dark)';

const ThemeContextProvider: FC<PropsWithChildren> = props => {
  const { defaultAlgorithm, darkAlgorithm } = theme;

  const { children } = props;

  const [isDarkTheme, setIsDarkTheme] = useState(() => {
    // first check if user has already set theme
    const localIsDarkTheme = localStorage.getItem(
      LOCALSTORAGE_IS_DARK_THEME_KEY,
    );
    if (localIsDarkTheme) {
      return JSON.parse(localIsDarkTheme);
    } else {
      // if not, check if browser/device has theme preference
      const lightMediaQuery = window.matchMedia(
        MEDIA_QUERY_PREFER_LIGHT_SCHEMA,
      ).matches;
      if (lightMediaQuery) return false;
    }
    // default to true
    return true;
  });

  // reflect state change to css change
  useEffect(() => {
    if (isDarkTheme) {
      localStorage.setItem(
        LOCALSTORAGE_IS_DARK_THEME_KEY,
        JSON.stringify(true),
      );
      document.body.classList.remove('light');
      document.body.classList.add('dark');
    } else {
      localStorage.setItem(
        LOCALSTORAGE_IS_DARK_THEME_KEY,
        JSON.stringify(false),
      );
      document.body.classList.remove('dark');
      document.body.classList.add('light');
    }
  }, [isDarkTheme]);

  // setup event listeners for browser media-query API prefer-color-schema change
  // probably excessive but it doesn't hurt
  useEffect(() => {
    const matchPreferLight = window.matchMedia(MEDIA_QUERY_PREFER_LIGHT_SCHEMA);
    const preferLightCallback = (e: MediaQueryListEvent) =>
      e.matches && setIsDarkTheme(false);
    matchPreferLight.addEventListener('change', preferLightCallback);

    const matchPreferDark = window.matchMedia(MEDIA_QUERY_PREFER_DARK_SCHEMA);
    const preferDarkCallback = (e: MediaQueryListEvent) =>
      e.matches && setIsDarkTheme(true);
    matchPreferDark.addEventListener('change', preferDarkCallback);

    return () => {
      matchPreferLight.removeEventListener('change', preferLightCallback);
      matchPreferDark.removeEventListener('change', preferDarkCallback);
    };
  }, []);

  return (
    <ThemeContext.Provider value={{ isDarkTheme, setIsDarkTheme }}>
      <StyleProvider layer>
        <ConfigProvider
          theme={{
            algorithm: isDarkTheme ? darkAlgorithm : defaultAlgorithm,
            hashed: false,
          }}
        >
          {children}
        </ConfigProvider>
      </StyleProvider>
    </ThemeContext.Provider>
  );
};

export default ThemeContextProvider;
