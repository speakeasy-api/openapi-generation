import {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useState,
} from 'react';
import { useRouter } from 'next/router';

type RouteProviderProps = {
  route: string;
  setRoute: (route: string) => void;
};
export const RouteContext = createContext<RouteProviderProps>({
  route: '/',
  setRoute: () => {},
});

export const useRoute = () => useContext(RouteContext).route;

export const RouteProvider = (props: { children: ReactNode }) => {
  const router = useRouter();
  const [route, setRoute] = useState('');

  useEffect(() => {
    if (router.isReady) {
      // This is strange looking, but at this point router.asPath is still the old route
      // But without the router.isReady check, window.location.pathname will also still be the old route
      setRoute(window.location.pathname);
    }
  }, [router.isReady]);

  const handleRouteChange = (newRoute: string) => {
    const currentBasePage = route.split('/').at(1);

    setRoute(newRoute);

    // If we're on the same root page, just update the URL so that we don't have to reload the page
    if (newRoute.startsWith('/' + currentBasePage)) {
      updateUrlShallow(window, newRoute);
    } else {
      // This causes the page to reload
      router.push(newRoute, newRoute, { scroll: false });
    }
  };

  return (
    <RouteContext.Provider value={{ route, setRoute: handleRouteChange }}>
      {props.children}
    </RouteContext.Provider>
  );
};

// We take in a window to ensure this is only called from where window is available
export const updateUrlShallow = (window: Window, url: string) => {
  window.history.replaceState(
    {
      ...window.history.state,
      as: url,
      url: url,
    },
    '',
    url,
  );
};
