// from https://github.com/novnc/noVNC/pull/1347

export default class Clipboard {
  private _target: HTMLElement;
  private _eventHandlers: {
    copy: (e: ClipboardEvent) => void;
    paste: (e: ClipboardEvent) => void;
  };
  onpaste: (v: string) => void;
  static isSupported: any = !!(navigator && navigator.clipboard);

  constructor(target: HTMLElement) {
    this._target = target;

    this._eventHandlers = {
      copy: this._handleCopy.bind(this),
      paste: this._handlePaste.bind(this),
    };

    // ===== EVENT HANDLERS =====

    this.onpaste = () => {};
  }

  // ===== PRIVATE METHODS =====

  _handleCopy(e: ClipboardEvent) {
    if (navigator.clipboard.writeText && e.clipboardData) {
      navigator.clipboard
        .writeText(e.clipboardData.getData('text/plain'))
        .catch(() => {
          /* Do nothing */
        });
    }
  }

  _handlePaste(e: ClipboardEvent) {
    if (navigator.clipboard.readText) {
      navigator.clipboard
        .readText()
        .then(this.onpaste)
        .catch(() => {
          /* Do nothing */
        });
    } else if (e.clipboardData) {
      this.onpaste(e.clipboardData.getData('text/plain'));
    }
  }

  // ===== PUBLIC METHODS =====

  grab() {
    if (!Clipboard.isSupported) return;
    this._target.addEventListener('copy', this._eventHandlers.copy);
    this._target.addEventListener('paste', this._eventHandlers.paste);
  }

  ungrab() {
    if (!Clipboard.isSupported) return;
    this._target.removeEventListener('copy', this._eventHandlers.copy);
    this._target.removeEventListener('paste', this._eventHandlers.paste);
  }
}
