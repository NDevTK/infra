import sinon from 'sinon';
import {assert} from 'chai';
import {CronTask} from './cron.js';

describe('cron', () => {
  beforeEach(() => {
    let count = 0;
    sinon.stub(window, 'setTimeout').callsFake((fn) => {
      // Execute the passed argument only twice to avoid being stuck in a loop.
      if (count < 2) {
        count += 1;
        fn();
      }
    });
  });

  afterEach(() => {
    window.setTimeout.restore();
  });

  it('calls task periodically', () => {
    const task = sinon.spy();
    const cronTask = new CronTask(task, 1234);

    // Make sure task is not called until the cron task has been started.
    assert.isFalse(task.called);

    cronTask.start();

    // task must have been called thrice: once whin the cron task was started,
    // and twice after being called by setTimeout.
    assert.isTrue(task.calledThrice);

    // setTimeout must have been called thrice, always with the given delay.
    assert.isTrue(window.setTimeout.calledThrice);
    assert.isTrue(window.setTimeout.alwaysCalledWith(sinon.match.any, 1234));
  });
});
