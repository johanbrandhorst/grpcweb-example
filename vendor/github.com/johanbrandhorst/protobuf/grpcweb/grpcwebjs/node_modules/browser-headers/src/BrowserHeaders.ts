import { normalizeName, normalizeValue, getHeaderValues, getHeaderKeys, splitHeaderValue } from "./util";
import { WindowHeaders } from "./WindowHeaders";

export interface Map<K, V> {
  clear(): void;
  delete(key: K): boolean;
  forEach(callbackfn: (value: V, index: K, map: Map<K, V>) => void, thisArg?: any): void;
  get(key: K): V | undefined;
  has(key: K): boolean;
  set(key: K, value?: V): this;
  readonly size: number;
}

interface MapConstructor {
  new (): Map<any, any>;
  new <K, V>(entries?: [K, V][]): Map<K, V>;
  readonly prototype: Map<any, any>;
}

declare const Map: MapConstructor;

// Declare that there is a global property named "Headers" - this might not be present at runtime
declare const Headers: any;

// BrowserHeaders is a wrapper class for Headers
export class BrowserHeaders {
  keyValueMap: {[key: string]: string[]};

  constructor(init: BrowserHeaders.ConstructorArg = {}, options: {splitValues: boolean} = { splitValues: false } ) {
    this.keyValueMap = {};

    if (init) {
      if (typeof Headers !== "undefined" && init instanceof Headers) {
        const keys = getHeaderKeys(init as WindowHeaders);
        keys.forEach(key => {
          const values = getHeaderValues(init as WindowHeaders, key);
          values.forEach(value => {
            if (options.splitValues) {
              this.append(key, splitHeaderValue(value));
            } else {
              this.append(key, value);
            }
          });
        });
      } else if (init instanceof BrowserHeaders) {
        (init as BrowserHeaders).forEach((key, values) => {
          this.append(key, values)
        });
      } else if (typeof Map !== "undefined" && init instanceof Map) {
        const asMap = init as BrowserHeaders.HeaderMap;
        asMap.forEach((value: string|string[], key: string) => {
          this.append(key, value);
        });
      } else if (typeof init === "string") {
        this.appendFromString(init);
      } else if (typeof init === "object") {
        Object.getOwnPropertyNames(init).forEach(key => {
          const asObject = init as BrowserHeaders.HeaderObject;
          const values = asObject[key];
          if (Array.isArray(values)) {
            values.forEach(value => {
              this.append(key, value);
            });
          } else {
            this.append(key, values);
          }
        });
      }
    }
  }

  appendFromString(str: string): void {
    const pairs = str.split("\r\n");
    for (let i = 0; i < pairs.length; i++) {
      const p = pairs[i];
      const index = p.indexOf(": ");
      if (index > 0) {
        const key = p.substring(0, index);
        const value = p.substring(index + 2);
        this.append(key, value);
      }
    }
  }

  // delete either the key (all values) or a specific value for a key
  delete(key: string, value?: string): void {
    const normalizedKey = normalizeName(key);
    if (value === undefined) {
      delete this.keyValueMap[normalizedKey];
    } else {
      const existing = this.keyValueMap[normalizedKey];
      if (existing) {
        const index = existing.indexOf(value);
        if (index >= 0) {
          existing.splice(index, 1);
        }
        if (existing.length === 0) {
          // The last value was removed - remove the key
          delete this.keyValueMap[normalizedKey];
        }
      }
    }
  }

  append(key: string, value: string | string[]): void {
    const normalizedKey = normalizeName(key);
    if (!Array.isArray(this.keyValueMap[normalizedKey])) {
      this.keyValueMap[normalizedKey] = [];
    }
    if (Array.isArray(value)) {
      value.forEach(arrayValue => {
        this.keyValueMap[normalizedKey].push(normalizeValue(arrayValue));
      });
    } else {
      this.keyValueMap[normalizedKey].push(normalizeValue(value));
    }
  }

  // set overrides all existing values for a key
  set(key: string, value: string | string[]): void {
    const normalizedKey = normalizeName(key);
    if (Array.isArray(value)) {
      const normalized: string[] = [];
      value.forEach(arrayValue => {
        normalized.push(normalizeValue(arrayValue));
      });
      this.keyValueMap[normalizedKey] = normalized;
    } else {
      this.keyValueMap[normalizedKey] = [normalizeValue(value)];
    }
  }

  has(key: string, value?: string): boolean {
    const keyArray = this.keyValueMap[normalizeName(key)];
    const keyExists = Array.isArray(keyArray);
    if (!keyExists) {
      return false;
    }
    if (value !== undefined) {
      const normalizedValue = normalizeValue(value);
      return keyArray.indexOf(normalizedValue) >= 0;
    } else {
      return true;
    }
  }

  get(key: string): string[] {
    const values = this.keyValueMap[normalizeName(key)];
    if (values !== undefined) {
      return values.concat();
    }
    return [];
  }

  // forEach iterates through the keys and calls the callback with the key and *all* of it's values as an array
  forEach(callback: (key: string, values: string[]) => void): void {
    Object.getOwnPropertyNames(this.keyValueMap)
      .forEach(key => {
        callback(key, this.keyValueMap[key]);
      }, this);
  }
  
  toHeaders(): WindowHeaders {
    if (typeof Headers !== "undefined") {
      const headers: WindowHeaders = new Headers();
      this.forEach((key, values) => {
        values.forEach(value => {
          headers.append(key, value);
        });
      });
      return headers;
    } else {
      throw new Error("Headers class is not defined");
    }
  }
}

export namespace BrowserHeaders {
  export type HeaderObject = {[key: string]: string|string[]};
  export type HeaderMap = Map<string, string|string[]>;
  export type ConstructorArg = HeaderObject | HeaderMap | BrowserHeaders | WindowHeaders | string;
}
