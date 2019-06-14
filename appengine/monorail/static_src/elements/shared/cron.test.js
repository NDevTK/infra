import sinon from 'sinon';
import {assert} from 'chai';
import {CronTask} from './cron.js';

let clock;

describe('cron', () => {
  beforeEach(() => {
    clock = sinon.useFakeTimers();
  });

  afterEach(() => {
    clock.restore();
  });

  it('calls task periodically', () => {
    const task = sinon.spy();
    const cronTask = new CronTask(task, 1000);

    // Make sure task is not called until the cron task has been started.
    assert.isFalse(task.called);

    cronTask.start();
    assert.isTrue(task.calledOnce);

    clock.tick(1000);
    assert.isTrue(task.calledTwice);

    clock.tick(1000);
    assert.isTrue(task.calledThrice);
  });
});
