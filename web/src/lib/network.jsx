import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import { api } from './api.js';

// État partagé entre les écrans : résumé, appareils, alertes.
// Rafraîchi toutes les 30 s, et toutes les 4 s pendant un scan.
const NetworkContext = createContext(null);

export function NetworkProvider({ children }) {
  const [state, setState] = useState({ summary: null, devices: [], alerts: [], loading: true, error: null });
  const [requested, setRequested] = useState(false);
  const scanningRef = useRef(false);

  const refresh = useCallback(async () => {
    try {
      const [summary, devices, alerts] = await Promise.all([api.summary(), api.devices(), api.alerts()]);
      scanningRef.current = summary.scanning;
      if (!summary.scanning) setRequested(false);
      setState({ summary, devices, alerts, loading: false, error: null });
    } catch (e) {
      setState((s) => ({ ...s, loading: false, error: e.message }));
    }
  }, []);

  useEffect(() => {
    refresh();
    let timer;
    const tick = () => {
      timer = setTimeout(async () => {
        await refresh();
        tick();
      }, scanningRef.current ? 4000 : 30000);
    };
    tick();
    return () => clearTimeout(timer);
  }, [refresh]);

  const requestScan = useCallback(async () => {
    setRequested(true);
    scanningRef.current = true;
    try {
      await api.requestScan();
    } catch (e) {
      setRequested(false);
      throw e;
    }
    refresh();
  }, [refresh]);

  // Mise à jour optimiste d'une alerte (lue / non lue) pour une interface réactive
  const setAlertRead = useCallback(async (id, read) => {
    setState((s) => ({ ...s, alerts: s.alerts.map((a) => (a.id === id ? { ...a, read } : a)) }));
    try {
      await api.setAlertRead(id, read);
    } finally {
      refresh();
    }
  }, [refresh]);

  const markAllRead = useCallback(async () => {
    setState((s) => ({ ...s, alerts: s.alerts.map((a) => ({ ...a, read: true })) }));
    try {
      await api.markAllRead();
    } finally {
      refresh();
    }
  }, [refresh]);

  const updateDevice = useCallback(async (ip, patch) => {
    await api.updateDevice(ip, patch);
    await refresh();
  }, [refresh]);

  const scanning = requested || !!state.summary?.scanning;

  return (
    <NetworkContext.Provider value={{ ...state, scanning, refresh, requestScan, setAlertRead, markAllRead, updateDevice }}>
      {children}
    </NetworkContext.Provider>
  );
}

export const useNetwork = () => useContext(NetworkContext);
