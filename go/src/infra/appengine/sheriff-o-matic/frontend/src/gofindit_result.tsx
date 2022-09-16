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
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';

interface GoFinditResultProps {
  result: GoFinditAnalyis[];
}

interface GoFinditAnalyis {
  analysis_id: string;
  heuristic_result: HeuristicAnalysis;
  culprits: Culprit[];
}

interface HeuristicAnalysis {
  suspects: HeuristicSuspect[];
}

interface HeuristicSuspect {
  reviewUrl: string;
  justification: string;
  score: number;
  confidence_level: number;
}

interface Culprit {
  review_url: string;
  review_title: string;
}

interface LuciBisectionCulpritSectionProps {
  culprits: Culprit[];
}

export const LuciBisectionCulpritSection = ({ culprits }: LuciBisectionCulpritSectionProps) => {
  return <>
    <Link href="https://luci-bisection.appspot.com" target='_blank' rel='noopener'>
      LUCI Bisection
    </Link>
    &nbsp; has identified the following CL(s) as culprit(s) of the failure:
    <List>
    {
      culprits.map((c) => (
        <ListItem>
          <Link
            href={c.review_url}
            target='_blank'
            rel='noopener'
            onClick={() => {
              ga('send', {
                hitType: 'event',
                eventCategory: 'GoFindit',
                eventAction: 'ClickCulpritLink',
                eventLabel: c.review_url,
                transport: 'beacon',
              });
            }}
          >
            {getCulpritDisplayUrl(c)}
          </Link>
        </ListItem>
      ))
    }
    </List>
  </>
}

export const GoFinditResult = (props: GoFinditResultProps) => {
  if (props.result == null) {
    return <></>
  }

  // If there is a culprit, display it
  const culprits = props.result.flatMap(r => r.culprits)
  if (culprits.length > 0) {
    return <LuciBisectionCulpritSection culprits={culprits}></LuciBisectionCulpritSection>
  }

  const suspects = props.result.flatMap(r => r.heuristic_result?.suspects)
  if (suspects.length == 0) {
    return <></>;
  }

  const goFinditTooltipTitle = "LUCI Bisection (http://luci-bisection.appspot.com) has identified the following CLs as suspects.\nThis was based on best-guess effort, so it may not be 100% accurate."
  return <>
    <TableContainer>
      <Table sx={{ maxWidth: "1000px", tableLayout: "fixed"}}>
        <TableHead>
          <TableRow>
            <TableCell align="left" sx={{ width: "350px"}}>
              LUCI Bisection Suspected CL
              {/* We dont use Tooltip from MUI here as the MUI tooltip is attached in <body> tag
                  so style cannot be applied. */}
              <span title={goFinditTooltipTitle}>
                <IconButton>
                  <HelpOutline></HelpOutline>
                </IconButton>
              </span>
            </TableCell>
            <TableCell align="left" sx={{ width: "60px"}}>
              Confidence
            </TableCell>
            <TableCell align="left">Justification</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {suspects.map((s) => (
            <TableRow>
              <TableCell align='left'>
                <Link
                  href={s.reviewUrl}
                  target='_blank'
                  rel='noopener'
                  onClick={() => {
                    ga('send', {
                      hitType: 'event',
                      eventCategory: 'LuciBisection',
                      eventAction: 'ClickSuspectLink',
                      eventLabel: s.reviewUrl,
                      transport: 'beacon',
                    });
                  }}
                >
                  {s.reviewUrl}
                </Link>
              </TableCell>
              <TableCell align="left">{confidenceText(s.confidence_level)}</TableCell>
              <TableCell align="left">
                <pre style={{ whiteSpace: 'pre-wrap' }} >{shortenJustification(s.justification)}</pre>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  </>
}

// Sometimes justification is too long if a CL touches many files.
// In such case we should shorten the justification to at most 3 lines
// and link to the detailed analysis if the sheriff wants to see the details
// (when the detail analysis page is ready).
function shortenJustification(justification: string) {
  const lines = justification.split("\n");
  if (lines.length < 4) {
    return justification
  }
  return lines.slice(0, 3).join("\n") + "\n...";
}

function confidenceText(confidenceLevel: number) {
  switch (confidenceLevel) {
    case 1:
      return "Low";
    case 2:
      return "Medium";
    case 3:
      return "High";
    default:
      return "N/A";
  }
}

function getCulpritDisplayUrl(c: Culprit) {
  if (c.review_title == "" || c.review_title == null) {
    return c.review_url
  }
  return c.review_title
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