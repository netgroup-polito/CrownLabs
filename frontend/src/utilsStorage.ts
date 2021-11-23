export enum StorageKeys {
  // Active Page Keys
  Active_View = 'ActivePageView',
  Active_Headers = 'ActivePageHeaders',
  Active_ID_WS = 'ActivePageIDWorkspace',
  Active_ID_T = 'ActivePageIDTemplate',
  // Dashboard Keys
  Dashboard_View = 'DashboardPageView',
  Dashboard_ID_T = 'DashboardPageTemplate',
}

class StorageValue {
  private key: StorageKeys;
  private defaultValue: string;
  private storageBackend: Storage;

  constructor(key: StorageKeys, defaultValue: string, storageBackend: Storage) {
    this.key = key;
    this.defaultValue = defaultValue;
    this.storageBackend = storageBackend;
  }

  public getKey(keyExtension = ''): string {
    return `${this.key}${keyExtension}`;
  }

  public getDefault(): string {
    return this.defaultValue;
  }

  public get(keyExtension = ''): string {
    return (
      this.storageBackend.getItem(this.getKey(keyExtension)) ||
      this.getDefault()
    );
  }

  public set(value: string, keyExtension = ''): void {
    this.storageBackend.setItem(this.getKey(keyExtension), value);
  }

  public isSet(keyExtension = ''): boolean {
    return this.storageBackend.getItem(this.getKey(keyExtension)) !== null;
  }

  public remove(keyExtension = ''): void {
    this.storageBackend.removeItem(this.getKey(keyExtension));
  }
}

export class LocalValue extends StorageValue {
  constructor(key: StorageKeys, defaultValue: string) {
    super(key, defaultValue, localStorage);
  }
}

export class SessionValue extends StorageValue {
  constructor(key: StorageKeys, defaultValue: string) {
    super(key, defaultValue, sessionStorage);
  }
}
