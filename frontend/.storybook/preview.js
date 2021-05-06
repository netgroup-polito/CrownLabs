// include tailwind classes in the storybook
import '../src/includeTailwind.css';
const TAILWIND_TEXT_MATCH = 'TAILWIND_STYLE_IMPORT';

export const parameters = {
  actions: { argTypesRegex: '^on[A-Z].*' },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
  },
};

// this functions is a logic to handle the case where there are multiple potential local css files
const findCorrectCssFile = async (cssLinkElements, match) => {
  for (let i = 0; i < cssLinkElements.length; i++) {
    const cssHref = cssLinkElements[i].href;
    try {
      const res = await fetch(cssHref);
      const text = await res.text();

      if (text.includes(match)) {
        cssLinkElements[i].remove();
        const styleElemToAppend = document.createElement('style');
        styleElemToAppend.innerText = text;
        document.head.appendChild(styleElemToAppend);
        console.info('TAILWIND CSS LOADED');
        break;
      }
    } catch (err) {
      if (i === cssLinkElements.length - 1) {
        console.error(`Last css href for tailwind tried was ${cssHref}`);
        throw new Error('NO CSS LINK FOUND FOR TAILWIND');
      } else {
        console.info(
          `Error when looking for tailwind css for path ${cssHref}, continuing to next link`,
          err
        );
      }
    }
  }
};

// this decorator is to prioritize the tailwind css over antd for each story
// by brutally moving the style tag of tailwind to the bottom of the head
const tailwindPrioritizerDec = Story => {
  if (process.env.NODE_ENV === 'development') {
    const styleTail = Array.from(
      document.querySelectorAll('style')
    ).find(elem => elem.innerText.includes(TAILWIND_TEXT_MATCH));

    styleTail.remove();
    document.head.appendChild(styleTail);
  } else {
    const potentialLinks = Array.from(
      document.querySelectorAll("link[rel='stylesheet'][href$='.css']")
    );

    findCorrectCssFile(potentialLinks, TAILWIND_TEXT_MATCH);
  }

  return <Story />;
};

export const decorators = [tailwindPrioritizerDec];
