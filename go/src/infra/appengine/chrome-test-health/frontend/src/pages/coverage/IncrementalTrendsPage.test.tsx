import { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import * as TrendsContext from '../../features/coverage/trends/TrendsContext';
import IncrementalTrendsPage from './IncrementalTrendsPage';

export function renderWithBrowserRouter(
    ui: ReactElement,
) {
  render(
      <BrowserRouter>
        {ui}
      </BrowserRouter>,
  );
}

describe('when rendering the IncrementalTrends Page', () => {
  // This is needed to allow us to modify window.location
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { assign: jest.fn() },
  });

  it('should pass in default values', async () => {
    const mockContext = jest.fn();
    jest.spyOn(TrendsContext, 'TrendsContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    renderWithBrowserRouter(<IncrementalTrendsPage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          platform: '',
          unitTestsOnly: false,
          paths: [],
          presets: [],
        }),
    );
  });

  it('should pass in url param values', async () => {
    const mockContext = jest.fn();
    jest.spyOn(TrendsContext, 'TrendsContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    window.location.search = 'https://localhost/?placeholder'+
    '=placeholder&paths=//a/b/&isTest=true';
    renderWithBrowserRouter(<IncrementalTrendsPage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          paths: ['//a/b/'],
          unitTestsOnly: true,
        }),
    );
  });
});
