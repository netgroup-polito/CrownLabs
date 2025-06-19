import unittest
from delete_stale_instances import InstanceExpiredDeleter
from datetime import datetime
from datetime import timedelta
import math


class TestDeleteCron(unittest.TestCase):

    def test_instance_is_expired(self):
        """
        Test the correct behavior of instance_is_expired function.
        :Two test cases both for True (expired date) and False (not expired date) cases.
        :threshold used for testing set to two days. can be changed to any value (in seconds).
        """
        # take now time
        now = datetime.now()
        # define a expiration threshold for testing
        threshold = 3600 * 24 * 2
        # define a delta time to test case expired
        expiration = threshold + 2
        # define expired time to test expired case
        expired_time = now - timedelta(seconds=expiration)
        # define not expired time to test not expired case
        notexpired_time = now
        # do the tests
        result_expired = InstanceExpiredDeleter.instance_is_expired(threshold, expired_time.strftime('%Y-%m-%dT%H:%M:%SZ'))
        # case expired so result should be True
        self.assertEqual(result_expired, True)
        result_notexpired = InstanceExpiredDeleter.instance_is_expired(
            threshold, notexpired_time.strftime('%Y-%m-%dT%H:%M:%SZ'))
        # case not Expired so result should be False
        self.assertEqual(result_notexpired, False)

    def test_instance_is_expired_infinite(self):
        """
        Test that the instance_is_expired function returns false when the lifespan is infinite.
        """

        expired = InstanceExpiredDeleter.instance_is_expired(
            math.inf, '2001-01-05T15:45:09Z')
        self.assertEqual(expired, False)

    def test_convert_to_life_span(self):
        """
        Test the correct behavior of convert_to_life_span function.
        :Six test cases: never policy, minutes case, hour case, day case and two wrong format cases.
        """
        never_string = "never"
        expected_never = math.inf
        minutes_string = "9m"
        expected_minutes = 9 * 60
        hours_string = "20h"
        expected_hours = 20 * 60 * 60
        days_string = "2d"
        expected_days = 2 * 24 * 60 * 60
        wrong_format1_string = "m10"
        wrong_format2_string = "25l"
        expected_wrong = None

        never = InstanceExpiredDeleter.convert_to_life_span(never_string)
        self.assertEqual(never, expected_never)
        minutes = InstanceExpiredDeleter.convert_to_life_span(minutes_string)
        self.assertEqual(minutes, expected_minutes)
        hours = InstanceExpiredDeleter.convert_to_life_span(hours_string)
        self.assertEqual(hours, expected_hours)
        days = InstanceExpiredDeleter.convert_to_life_span(days_string)
        self.assertEqual(days, expected_days)
        wrong_format1 = InstanceExpiredDeleter.convert_to_life_span(wrong_format1_string)
        self.assertEqual(wrong_format1, expected_wrong)
        wrong_format2 = InstanceExpiredDeleter.convert_to_life_span(wrong_format2_string)
        self.assertEqual(wrong_format2, expected_wrong)


if __name__ == '__main__':
    unittest.main()
