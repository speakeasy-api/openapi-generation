import React, {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { RouteContext } from '@/src/components/routeProvider';
import { getPagesUnderRoute } from 'nextra/context';
import { useRouter } from 'next/router';

export const MultiPageContext = createContext(false);
export const ScrollContext = createContext<{
  headingToPosition: Record<string, HeadingPosition>;
  currentHeading: string;
  visibleHeadings: string[];
  upsertHeading: (
    route: string,
    heading: string,
    elem: HTMLHeadingElement,
    position: number,
  ) => void;
  scrollTo: (route: string) => void;
  setPage: (route: string) => void;
}>({
  headingToPosition: {},
  currentHeading: '',
  visibleHeadings: [],
  upsertHeading: () => {},
  scrollTo: () => {},
  setPage: () => {},
});

export const useSetPage = () => useContext(ScrollContext).setPage;

type HeadingPosition = {
  elem: HTMLHeadingElement;
  position: number;
};

// Used to change the route a bit before the heading is at the top of the page
const headingOffset = -200;

export const ScrollManager = (props: {
  children: ReactNode;
}): React.ReactElement => {
  const isMultipage = useContext(MultiPageContext);
  const { route, setRoute } = useContext(RouteContext);
  const router = useRouter();

  const [initialScrollTarget, setInitialScrollTarget] = useState<string>();
  const [initialScrollDone, setInitialScrollDone] = useState(false);

  const rootPage = useMemo(
    () => (isMultipage ? route?.split('/').at(1) ?? '' : ''),
    [route],
  );

  const reset = () => {
    setHeadingToPosition({});
    setInitialScrollDone(false);
    setInitialScrollTarget(undefined);
  };

  const setPage = async (route: string) => {
    reset();
    setRoute(route);
    handleInitialScroll(route);
  };

  // If the route exists, set it as the scroll target. Otherwise, redirect to the root page.
  const handleInitialScroll = (route: string) => {
    if (pageExists(route)) {
      setInitialScrollTarget(route);
    } else {
      const language = route.split('/')[1];
      router.replace(`/${language}/client_sdks`);
      setInitialScrollDone(true);
    }
  };

  const [headingToPosition, setHeadingToPosition] = useState<
    Record<string, HeadingPosition>
  >({});
  const upsertHeading = (
    route: string,
    heading: string,
    elem: HTMLHeadingElement,
    position: number,
  ) => {
    setHeadingToPosition((currentValues) => {
      position = position + headingOffset;

      // If there are multiple headings in a section, we want to keep only the topmost one.
      // As a result, clicking the link in the sidebar will correctly scroll to the top of the section
      const current = currentValues[route];

      if (current && current.elem !== elem && position > current.position) {
        route += `#${heading}`;
      }

      const newValues = {
        ...currentValues,
        [route]: {
          elem,
          position,
        },
      };

      // This is necessary because sometimes the page loads out of order when we do client-side redirects
      return Object.fromEntries(
        Object.entries(newValues).filter(([k, _]) =>
          k.startsWith(`/${rootPage}`),
        ),
      );
    });
  };

  const [closestHeading, setClosestHeading] = useState<string>('/' + rootPage);
  const [visibleHeadings, setVisibleHeadings] = useState<string[]>([
    closestHeading,
  ]);

  /**
   * This is responsible for setting the route in the URL to the closest heading when the user scrolls.
   * This is memoized so that it can be removed when the route changes (otherwise it prevents scrolling to the desired heading)
   */
  const scroll = useMemo(
    () => () => {
      const entries = Object.entries(headingToPosition);

      const visible = entries
        .filter(
          ([_, { position }]) =>
            window.scrollY < position &&
            position < window.scrollY + window.innerHeight,
        )
        .map(([route]) => route);

      // Find the first heading that is below the current scroll position
      const nextIndex = entries.findIndex(
        ([_, { position }]) => position > window.scrollY,
      );

      // The current heading is the one before that
      const currentIndex =
        nextIndex === -1
          ? entries.length - 1
          : nextIndex - 1 >= 0
          ? nextIndex - 1
          : 0;
      const closest = entries[currentIndex]?.[0];

      setClosestHeading(closest);
      setVisibleHeadings([closest, ...visible]);
    },
    [headingToPosition],
  );

  useEffect(() => {
    window.addEventListener('scroll', scroll, false);

    return () => window.removeEventListener('scroll', scroll, false);
  }, [scroll]);

  useEffect(() => {
    if (
      closestHeading &&
      initialScrollDone &&
      closestHeading.startsWith(`/${rootPage}`) // Make sure we haven't changed pages. Without this, we might overwrite the new route
    ) {
      setRoute(closestHeading);
    }
  }, [closestHeading]);

  // Scrolls the page to the location of the target heading
  const scrollTo = useMemo(
    () => (route: string) => {
      if (headingToPosition[route]) {
        document.addEventListener(
          'scrollend',
          () => {
            setClosestHeading(route);
          },
          { once: true },
        );

        // Scroll down a bit further than the heading so that it lines up right at the top
        window.scrollTo({ top: headingToPosition[route].position + 100 });
      }
    },
    [headingToPosition],
  );

  /**
   * On initial page load, set the heading to scroll to
   * This enables linking to a specific section
   * We don't want to run this every time the slug changes since we change it as the user scrolls
   */
  useEffect(() => {
    // At first, the slug is simply /[...rest], so wait til it properly pulls in the URL
    if (route && !initialScrollTarget) {
      handleInitialScroll(route);
    }
  }, [route]);

  // Once the initial scroll target is set and we know where that heading is, scroll to it
  // Do this every time the heading location changes until it stabilizes. This is necessary because the heading
  // will change positions on the page a few different times as the page loads. We want to scroll to it every time
  // it changes to reduce the perception of lagginess.
  useEffect(() => {
    let t: NodeJS.Timeout;
    if (
      initialScrollTarget &&
      headingToPosition[initialScrollTarget] &&
      !initialScrollDone
    ) {
      scrollTo(initialScrollTarget);
      t = setTimeout(() => {
        setInitialScrollDone(true);
      }, 50);
    }

    return () => clearTimeout(t);
  }, [initialScrollTarget && headingToPosition[initialScrollTarget]]);

  return (
    <ScrollContext.Provider
      value={{
        headingToPosition,
        upsertHeading,
        currentHeading: closestHeading,
        visibleHeadings,
        scrollTo,
        setPage,
      }}
    >
      {props.children}
    </ScrollContext.Provider>
  );
};

const pageExists = (route: string) => {
  // getPagesUnderRoute returns an empty array regardless of whether the page exists, so we instead get the
  // parent's children and check that the route exists there.
  const parentRoute = route.split('/').slice(0, -1).join('/');
  const parentsChildren = getPagesUnderRoute(parentRoute);
  return parentsChildren.some((child) => child.route === route);
};
