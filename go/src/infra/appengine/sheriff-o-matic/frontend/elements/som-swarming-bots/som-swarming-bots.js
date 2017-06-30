(function() {
'use strict';

Polymer({
  is: 'som-swarming-bots',

  properties: {
    bots: {
      type: Object,
      value: function() {
        return {
        }
      },
    },
    _hideDeadBots: {
      type: Boolean,
      computed: '_computeHideBots(bots.dead)',
      value: true,
    },
    _hideErrors: {
      type: Boolean,
      computed: '_computeHideBots(bots.errors)',
      value: true,
    },
    _hideQuarantinedBots: {
      type: Boolean,
      computed: '_computeHideBots(bots.quarantined)',
      value: true,
    },
    _toggleDeadIcon: {
      type: String,
      value: 'remove',
    },
    _toggleQuarantinedIcon: {
      type: String,
      value: 'remove',
    },
    _toggleSectionIcon: {
      type: String,
      computed: '_computeToggleSectionIcon(_opened)',
    },
    _opened: {
      type: Boolean,
      value: true,
    },
  },

  toggleDead: function() {
    this.$.deadBotsList.toggle();

    this._toggleDeadIcon = this._computeIcon(this.$.deadBotsList.opened);
  },

  toggleQuarantined: function() {
    this.$.quarantinedBotsList.toggle();

    this._toggleQuarantinedIcon =
        this._computeIcon(this.$.quarantinedBotsList.opened);
  },

  _toggleSection: function() {
    this._opened = !this._opened;
  },

  _computeHideBots: function(bots) {
    return !bots || bots.length == 0;
  },

  _collapseAll: function() {
    let deadBots = this.$.deadBotsList;
    let quarantinedBots = this.$.quarantinedBotsList;

    deadBots.opened = false;
    quarantinedBots.opened = false;

    this._toggleDeadIcon = this._computeIcon(this.$.deadBotsList.opened);
    this._toggleQuarantinedIcon =
        this._computeIcon(this.$.quarantinedBotsList.opened);
  },

  _expandAll: function() {
    this.$.deadBotsList.opened = true;
    this.$.quarantinedBotsList.opened = true;

    this._toggleDeadIcon = this._computeIcon(this.$.deadBotsList.opened);
    this._toggleQuarantinedIcon =
        this._computeIcon(this.$.quarantinedBotsList.opened);

    this._opened = true;
  },

  _computeToggleSectionIcon(opened) {
    return opened ? 'unfold-less' : 'unfold-more';
  },

  _computeIcon(opened) {
    return opened ? 'remove' : 'add';
  },
});
})();
