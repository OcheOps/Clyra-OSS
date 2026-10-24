import { openDB, type IDBPDatabase } from 'idb';

interface DiagramDB {
  projects: {
    key: string;
    value: {
      id: string;
      userId: string;
      name: string;
      diagram: unknown;
      createdAt: number;
      updatedAt: number;
    };
    indexes: {
      'by-updated': number;
      'by-user': string;
    };
  };
}

const DB_NAME = 'clyra-app';
const DB_VERSION = 2;

let dbPromise: Promise<IDBPDatabase<DiagramDB>> | null = null;

export function getDB(): Promise<IDBPDatabase<DiagramDB>> {
  if (!dbPromise) {
    dbPromise = openDB<DiagramDB>(DB_NAME, DB_VERSION, {
      upgrade(db, oldVersion) {
        // Fresh install or upgrading from v1
        if (oldVersion < 1) {
          const store = db.createObjectStore('projects', { keyPath: 'id' });
          store.createIndex('by-updated', 'updatedAt');
          store.createIndex('by-user', 'userId');
        } else if (oldVersion < 2) {
          // Upgrading from v1: add the by-user index
          const tx = (db as unknown as { transaction: (name: string, mode: string) => { objectStore: (name: string) => IDBObjectStore } })
            .transaction('projects', 'readwrite');
          const store = tx.objectStore('projects');
          if (!store.indexNames.contains('by-user')) {
            store.createIndex('by-user', 'userId');
          }
        }
      },
    });
  }
  return dbPromise;
}
