"""ANTLR3 runtime package"""

# begin[licence]
#
# [The "BSD licence"]
# Copyright (c) 2005-2008 Terence Parr
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions
# are met:
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
# 3. The name of the author may not be used to endorse or promote products
#    derived from this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
# IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
# OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
# IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
# INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
# NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
# DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
# THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
#
# end[licensc]

from antlr3.constants import EOF
from antlr3.exceptions import NoViableAltException, BacktrackingFailed


class DFA(object):
    """@brief A DFA implemented as a set of transition tables.

    Any state that has a semantic predicate edge is special; those states
    are generated with if-then-else structures in a specialStateTransition()
    which is generated by cyclicDFA template.

    """

    def __init__(
        self,
        recognizer, decisionNumber,
        eot, eof, min, max, accept, special, transition
        ):
        ## Which recognizer encloses this DFA?  Needed to check backtracking
        self.recognizer = recognizer

        self.decisionNumber = decisionNumber
        self.eot = eot
        self.eof = eof
        self.min = min
        self.max = max
        self.accept = accept
        self.special = special
        self.transition = transition


    def predict(self, input):
        """
        From the input stream, predict what alternative will succeed
	using this DFA (representing the covering regular approximation
	to the underlying CFL).  Return an alternative number 1..n.  Throw
	 an exception upon error.
	"""
        mark = input.mark()
        s = 0 # we always start at s0
        try:
            for _ in xrange(50000):
                #print "***Current state = %d" % s

                specialState = self.special[s]
                if specialState >= 0:
                    #print "is special"
                    s = self.specialStateTransition(specialState, input)
                    if s == -1:
                        self.noViableAlt(s, input)
                        return 0
                    input.consume()
                    continue

                if self.accept[s] >= 1:
                    #print "accept state for alt %d" % self.accept[s]
                    return self.accept[s]

                # look for a normal char transition
                c = input.LA(1)

                #print "LA = %d (%r)" % (c, unichr(c) if c >= 0 else 'EOF')
                #print "range = %d..%d" % (self.min[s], self.max[s])

                if c >= self.min[s] and c <= self.max[s]:
                    # move to next state
                    snext = self.transition[s][c-self.min[s]]
                    #print "in range, next state = %d" % snext

                    if snext < 0:
                        #print "not a normal transition"
                        # was in range but not a normal transition
                        # must check EOT, which is like the else clause.
                        # eot[s]>=0 indicates that an EOT edge goes to another
                        # state.
                        if self.eot[s] >= 0: # EOT Transition to accept state?
                            #print "EOT trans to accept state %d" % self.eot[s]

                            s = self.eot[s]
                            input.consume()
                            # TODO: I had this as return accept[eot[s]]
                            # which assumed here that the EOT edge always
                            # went to an accept...faster to do this, but
                            # what about predicated edges coming from EOT
                            # target?
                            continue

                        #print "no viable alt"
                        self.noViableAlt(s, input)
                        return 0

                    s = snext
                    input.consume()
                    continue

                if self.eot[s] >= 0:
                    #print "EOT to %d" % self.eot[s]

                    s = self.eot[s]
                    input.consume()
                    continue

                # EOF Transition to accept state?
                if c == EOF and self.eof[s] >= 0:
                    #print "EOF Transition to accept state %d" \
                    #  % self.accept[self.eof[s]]
                    return self.accept[self.eof[s]]

                # not in range and not EOF/EOT, must be invalid symbol
                self.noViableAlt(s, input)
                return 0

            else:
                raise RuntimeError("DFA bang!")

        finally:
            input.rewind(mark)


    def noViableAlt(self, s, input):
        if self.recognizer._state.backtracking > 0:
            raise BacktrackingFailed

        nvae = NoViableAltException(
            self.getDescription(),
            self.decisionNumber,
            s,
            input
            )

        self.error(nvae)
        raise nvae


    def error(self, nvae):
        """A hook for debugging interface"""
        pass


    def specialStateTransition(self, s, input):
        return -1


    def getDescription(self):
        return "n/a"


##     def specialTransition(self, state, symbol):
##         return 0


    def unpack(cls, string):
        """@brief Unpack the runlength encoded table data.

        Terence implemented packed table initializers, because Java has a
        size restriction on .class files and the lookup tables can grow
        pretty large. The generated JavaLexer.java of the Java.g example
        would be about 15MB with uncompressed array initializers.

        Python does not have any size restrictions, but the compilation of
        such large source files seems to be pretty memory hungry. The memory
        consumption of the python process grew to >1.5GB when importing a
        15MB lexer, eating all my swap space and I was to impacient to see,
        if it could finish at all. With packed initializers that are unpacked
        at import time of the lexer module, everything works like a charm.

        """

        ret = []
        for i in range(len(string) / 2):
            (n, v) = ord(string[i*2]), ord(string[i*2+1])

            # Is there a bitwise operation to do this?
            if v == 0xFFFF:
                v = -1

            ret += [v] * n

        return ret

    unpack = classmethod(unpack)
