const ogcv = getConfigVar;
getConfigVar = (a, b) => {
  if (a === 'autoconnect') return true;
  return ogcv(a, b);
}
