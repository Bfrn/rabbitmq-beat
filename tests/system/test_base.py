from rabbitmq-beat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Rabbitmq-beat normally
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/*"
        )

        rabbitmq-beat_proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("rabbitmq-beat is running"))
        exit_code = rabbitmq-beat_proc.kill_and_wait()
        assert exit_code == 0
