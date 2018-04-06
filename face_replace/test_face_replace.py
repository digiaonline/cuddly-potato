#!/usr/bin/env python
"""
Unit tests
"""
from __future__ import print_function, unicode_literals
import os
import shutil
import tempfile
import time
import unittest

import face_replace

X = 10
Y = 10
W = 100
H = 100


class TestImagePaths(unittest.TestCase):

    def test_filename(self):
        # Arrange
        filename = "README.md"

        # Act
        output = face_replace.image_paths(filename)

        # Assert
        self.assertEqual(output, ["README.md"])

    def test_spec(self):
        # Arrange
        spec = "*.md"

        # Act
        output = face_replace.image_paths(spec)

        # Assert
        self.assertEqual(output, ["README.md"])


class TestImagePathsDir(unittest.TestCase):

    def setUp(self):
        # Create a new, empty directory
        self.dir = os.path.join(tempfile.gettempdir() + str(time.time()))
        os.mkdir(self.dir)

        # Add some dummy files
        self.file1 = os.path.join(self.dir, 'file1.png')
        self.file2 = os.path.join(self.dir, 'file2.png')
        open(self.file1, 'a').close()
        open(self.file2, 'a').close()

    def tearDown(self):
        shutil.rmtree(self.dir)

    def test_dir(self):
        # Act
        output = face_replace.image_paths(self.dir)

        # Assert
        self.assertEqual(len(output), 2)
        self.assertIn(self.file1, output)
        self.assertIn(self.file2, output)


class TestImagePathsEmpty(unittest.TestCase):

    def setUp(self):
        # Create a new, empty directory
        self.dir = os.path.join(tempfile.gettempdir() + str(time.time()))
        os.mkdir(self.dir)

    def tearDown(self):
        os.rmdir(self.dir)

    def test_empty_dir(self):
        # Act
        output = face_replace.image_paths(self.dir)

        # Assert
        self.assertEqual(output, [])


if __name__ == '__main__':
    unittest.main()

# End of file
