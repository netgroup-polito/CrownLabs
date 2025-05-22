import { createContext } from 'react';

interface IThemeContext {
  isDarkTheme: boolean;
  setIsDarkTheme: React.Dispatch<React.SetStateAction<boolean>>;
}

export const ThemeContext = createContext<IThemeContext>({
  isDarkTheme: false,
  setIsDarkTheme: () => {},
});
