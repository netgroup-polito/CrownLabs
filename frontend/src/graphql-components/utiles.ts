function beautifyGqlResponse(obj: any): any {
  if (typeof obj !== 'object' || obj === null) return obj;

  if (Array.isArray(obj)) {
    let newArray: Array<any> = [];
    obj.forEach((e: any) => newArray.push(beautifyGqlResponse(e)));
    return newArray;
  }

  let newObj: any = {};
  const keys: any = Object.keys(obj);
  keys.forEach((key: any) => {
    let tmp: any = beautifyGqlResponse(obj[key]);
    if (typeof tmp !== 'object' || tmp === null || Array.isArray(tmp)) {
      newObj[key] = tmp;
    } else {
      const keysTmp: any = Object.keys(tmp);
      keysTmp.forEach((keyTmp: any) => {
        newObj[keyTmp] = tmp[keyTmp];
      });
    }
  });
  return newObj;
}

export { beautifyGqlResponse };
