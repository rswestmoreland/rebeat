from rebeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Rebeat normally
        """
        self.render_config_template(
                path=os.path.abspath(self.working_dir) + "/log/*"
        )

        rebeat_proc = self.start_beat()
        self.wait_until( lambda: self.log_contains("rebeat is running"))
        exit_code = rebeat_proc.kill_and_wait()
        assert exit_code == 0
