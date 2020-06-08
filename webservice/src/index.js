import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';
import './App.css';
/* Since we do not have a index.html template, we create a main div where all our object will be drawn */
const main = document.createElement('div');
document.body.appendChild(main);
main.setAttribute('id', 'main');
// only place where to disable .jsx extension because it is root file
// eslint-disable-next-line react/jsx-filename-extension
ReactDOM.render(<App />, main);
