import React from 'react';

import { render } from 'react-dom';
import { CacheProvider } from '@emotion/react';
import createCache, { EmotionCache } from '@emotion/cache';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import HelpOutline from '@mui/icons-material/HelpOutline';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Link from '@mui/material/Link';

interface GoFinditResultProps {
  result: GoFinditAnalyis[];
}

interface GoFinditAnalyis {
  analysis_id: string;
  heuristic_result: HeuristicAnalysis;
}

interface HeuristicAnalysis {
  suspects: HeuristicSuspect[];
}

interface HeuristicSuspect {
  reviewUrl: string;
  justification: string;
  score: number;
}

export const GoFinditResult = (props: GoFinditResultProps) => {
  if (props.result == null) {
    return <></>
  }
  const suspects = props.result.flatMap(r => r.heuristic_result?.suspects)
  if (suspects.length == 0) {
    return <></>;
  }

  const goFinditTooltipTitle = "GoFindit (http://go/gofindit) has identified the following CLs as suspects.\nThis was based on best-guess effort, so it may not be 100% accurate."
  return <>
    <TableContainer>
      <Table sx={{ maxWidth: "1000px", tableLayout: "fixed"}}>
        <TableHead>
          <TableRow>
            <TableCell align="left" sx={{ width: "350px"}}>
              GoFindit Suspected CL
              {/* We dont use Tooltip from MUI here as the MUI tooltip is attached in <body> tag
                  so style cannot be applied. */}
              <span title={goFinditTooltipTitle}>
                <IconButton>
                  <HelpOutline></HelpOutline>
                </IconButton>
              </span>
            </TableCell>
            <TableCell align="left" sx={{ width: "50px"}}>
              Score
            </TableCell>
            <TableCell align="left">Justification</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {suspects.map((s) => (
            <TableRow>
              <TableCell align="left">
                <Link href={s.reviewUrl} target="_blank" rel="noopener">{s.reviewUrl}</Link>
              </TableCell>
              <TableCell align="left">{s.score}</TableCell>
              <TableCell align="left">
                <pre style={{ whiteSpace: 'pre-wrap' }} >{s.justification}</pre>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  </>
}

export class SomGoFinditResult extends HTMLElement {
  cache: EmotionCache;
  child: HTMLDivElement;
  props: GoFinditResultProps = {
    result: [],
  };

  constructor() {
    super();
    const root = this.attachShadow({ mode: 'open' });
    const parent = document.createElement('div');
    this.child = document.createElement('div');
    root.appendChild(parent).appendChild(this.child);
    this.cache = createCache({
      key: 'gofindit-result',
      container: parent,
    });
  }

  connectedCallback() {
    this.render();
  }

  set result(result: GoFinditAnalyis[]) {
    this.props.result = result;
    this.render();
  }

  render() {
    if (!this.isConnected) {
      return;
    }
    render(
      <CacheProvider value={this.cache}>
        <GoFinditResult {...this.props} />
      </CacheProvider>,
      this.child
    );
  }
}

customElements.define('som-gofindit-result', SomGoFinditResult);