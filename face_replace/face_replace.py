#!/usr/bin/env python
"""
From:
https://docs.opencv.org/3.3.0/d7/d8b/tutorial_py_face_detection.html
"""
from __future__ import print_function, unicode_literals
import argparse
import cv2
# import numpy as np
import os
import random
import sys


CRISUS = [
    "/Users/hugo/github/chrisify/faces/risuhead.png",
    "/Users/hugo/github/chrisify/faces/risuhead2.png",
    "/Users/hugo/github/chrisify/faces/risuhead3.png",
]


def image_resize(image, width=None, height=None, inter=cv2.INTER_AREA):
    # initialize the dimensions of the image to be resized and
    # grab the image size
    dim = None
    (h, w) = image.shape[:2]

    # if both the width and height are None, then return the
    # original image
    if width is None and height is None:
        return image

    # check to see if the width is None
    if width is None:
        # calculate the ratio of the height and construct the
        # dimensions
        r = height / float(h)
        dim = (int(w * r), height)

    # otherwise, the height is None
    else:
        # calculate the ratio of the width and construct the
        # dimensions
        r = width / float(w)
        dim = (width, int(h * r))

    # resize the image
    resized = cv2.resize(image, dim, interpolation=inter)

    # return the resized image
    return resized


def detect(infile, face_cascade_path, eye_cascade_path):

    face_cascade = cv2.CascadeClassifier(face_cascade_path)
    eye_cascade = cv2.CascadeClassifier(eye_cascade_path)
    l_img = cv2.imread(infile)
    gray = cv2.cvtColor(l_img, cv2.COLOR_BGR2GRAY)

    faces = face_cascade.detectMultiScale(gray, 1.3, 5)
    print(faces)
    for (x, y, w, h) in faces:
        cv2.rectangle(l_img, (x, y), (x+w, y+h), (255, 0, 0), 2)
        roi_gray = gray[y:y+h, x:x+w]
        roi_color = l_img[y:y+h, x:x+w]
        eyes = eye_cascade.detectMultiScale(roi_gray)
        for (ex, ey, ew, eh) in eyes:
            cv2.rectangle(roi_color, (ex, ey), (ex+ew, ey+eh), (0, 255, 0), 2)

        # Load a Crisu head, resize it and paste it
        s_img = cv2.imread(random.choice(CRISUS), -1)

        # Because we're detecting faces, not heads, but pasting heads,
        # we need to move the head up and left a bit, and make it bigger
        # to cover the face
        w_embiggen_factor = w*0.08
        h_embiggen_factor = h*0.3
        x = x - int(w_embiggen_factor)
        w = w + int(w_embiggen_factor)
        y = y - int(h_embiggen_factor)
        h = h + int(h_embiggen_factor)

        # Resize maintaining aspect ratio
        # s_img = image_resize(s_img, height=h)
        s_img = image_resize(s_img, width=w)

        cv2.rectangle(l_img, (x, y), (x + w, y + h), (0, 0, 255), 2)

        y1, y2 = y, y + s_img.shape[0]
        x1, x2 = x, x + s_img.shape[1]

        # Paste the new face, taking care of transparency
        alpha_s = s_img[:, :, 3] / 255.0
        alpha_l = 1.0 - alpha_s

        for c in range(0, 3):
            l_img[y1:y2, x1:x2, c] = (alpha_s * s_img[:, :, c] +
                                      alpha_l * l_img[y1:y2, x1:x2, c])

    cv2.imshow('img', l_img)
    cv2.waitKey(0)
    cv2.destroyAllWindows()
    sys.stdout.write(l_img.tostring())


if __name__ == '__main__':

    parser = argparse.ArgumentParser(
        description='TODO '
        'Requires OpenCV.',
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument(
        'infile',
        help='Image to mess up')
    parser.add_argument(
        '-cp', '--cascade-path',
        default='/usr/local/Cellar/opencv@2/2.4.13.5/share/OpenCV/'
                'haarcascades',
        help='Haar cascade file')
    parser.add_argument(
        '-fc', '--face-cascade',
        default='haarcascade_frontalface_alt.xml',
        help='Haar cascade file')
    parser.add_argument(
        '-ec', '--eye-cascade',
        default='haarcascade_eye.xml',
        help='Haar cascade file')
    parser.add_argument(
        '-s', '--show', action='store_true',
        help='Show detected image with box')

    args = parser.parse_args()
    # print(args)

    detect(args.infile,
           os.path.join(args.cascade_path, args.face_cascade),
           os.path.join(args.cascade_path, args.eye_cascade))
