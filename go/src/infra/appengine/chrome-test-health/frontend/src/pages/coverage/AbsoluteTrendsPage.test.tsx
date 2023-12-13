import { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import * as TrendsContext from '../../features/coverage/trends/TrendsContext';
import AbsoluteTrendsPage from './AbsoluteTrendsPage';

export function renderWithBrowserRouter(
    ui: ReactElement,
) {
  render(
      <BrowserRouter>
        {ui}
      </BrowserRouter>,
  );
}

describe('when rendering the AbsoluteTrends Page', () => {
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
    renderWithBrowserRouter(<AbsoluteTrendsPage/>);
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
    '=placeholder&paths=//a/b/&plat=linux&isTest=true';
    renderWithBrowserRouter(<AbsoluteTrendsPage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          paths: ['//a/b/'],
          platform: 'linux',
          unitTestsOnly: true,
        }),
    );
  });
});
