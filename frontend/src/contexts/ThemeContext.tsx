import {
  createContext,
  FC,
  PropsWithChildren,
  useEffect,
  useState,
} from 'react';

const LOCALSTORAGE_IS_DARL_THEME_KEY = 'REACT_APP_IS_DARK_THEME';
const MEDIA_QUERY_PREFER_LIGHT_SCHEMA = '(prefers-color-scheme: light)';
const MEDIA_QUERY_PREFER_DARK_SCHEMA = '(prefers-color-scheme: dark)';

interface IThemeContext {
  isDarkTheme: boolean;
  setIsDarkTheme: React.Dispatch<React.SetStateAction<boolean>>;
}

export const ThemeContext = createContext<IThemeContext>({
  isDarkTheme: false,
  setIsDarkTheme: () => {},
});

const ThemeContextProvider: FC<PropsWithChildren<{}>> = props => {
  const { children } = props;

  const [isDarkTheme, setIsDarkTheme] = useState(() => {
    // first check if user has already set theme
    const localIsDarkTheme = localStorage.getItem(
      LOCALSTORAGE_IS_DARL_THEME_KEY
    );
    if (localIsDarkTheme) {
      return JSON.parse(localIsDarkTheme);
    } else {
      // if not, check if browser/device has theme preference
      const lightMediaQuery = window.matchMedia(MEDIA_QUERY_PREFER_LIGHT_SCHEMA)
        .matches;
      if (lightMediaQuery) return false;
    }
    // default to true
    return true;
  });

  // reflect state change to css change
  useEffect(() => {
    if (isDarkTheme) {
      localStorage.setItem(
        LOCALSTORAGE_IS_DARL_THEME_KEY,
        JSON.stringify(true)
      );
      document.body.classList.remove('light');
    } else {
      localStorage.setItem(
        LOCALSTORAGE_IS_DARL_THEME_KEY,
        JSON.stringify(false)
      );
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
      {children}
    </ThemeContext.Provider>
  );
};

export default ThemeContextProvider;
