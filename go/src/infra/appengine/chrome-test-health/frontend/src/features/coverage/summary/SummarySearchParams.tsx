import { useContext, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ComponentContext, updateComponentsUrl } from '../../components/ComponentContext';
import { SummaryContext } from './SummaryContext';
import { Params } from './LoadSummary';

export const REVISION = 'revision';
export const UNIT_TESTS_ONLY = 'unitTestsOnly';
export const PLATFORM = 'platform';

function createSearchParams(components: string[], params: Params) {
  const search = new URLSearchParams();
  updateComponentsUrl(components, search);
  search.set(REVISION, params.revision);
  search.set(UNIT_TESTS_ONLY, `${params.unitTestsOnly}`);
  search.set(PLATFORM, params.platform);
  return search;
}


function SummarySearchParams() {
  const { params } = useContext(SummaryContext);
  const { components } = useContext(ComponentContext);
  const [, setSearchParams] = useSearchParams();

  useEffect(() => {
    setSearchParams(createSearchParams(components, params));
  }, [setSearchParams, params, components]);

  return (<></>);
}

export default SummarySearchParams;
