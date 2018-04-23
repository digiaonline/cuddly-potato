#!/usr/bin/env python
"""
Find faces in an input image and replace them with other random faces

https://docs.opencv.org/3.3.0/d7/d8b/tutorial_py_face_detection.html
"""
from __future__ import print_function, unicode_literals
import argparse
import cv2
import glob
import os
import random
import sys


def image_resize(image, width=None, height=None, inter=cv2.INTER_AREA):
    # Initialise the dimensions of the image to be resized and
    # grab the image size
    dim = None
    (h, w) = image.shape[:2]

    # If both the width and height are None, then return the original image
    if width is None and height is None:
        return image

    # Check to see if the width is None
    if width is None:
        # Calculate the ratio of the height and construct the dimensions
        r = height / float(h)
        dim = (int(w * r), height)

    # Otherwise, the height is None
    else:
        # Calculate the ratio of the width and construct the dimensions
        r = width / float(w)
        dim = (width, int(h * r))

    # Resize the image
    resized = cv2.resize(image, dim, interpolation=inter)

    # Return the resized image
    return resized


def paste_image(large, small, x1, y1, x2, y2):
    """Paste the small image into the large image at these coordinates,
    taking care of transparency
    """
    # Crop any bits of small extending beyond large's edges
    if y2 > large.shape[0]:
        y2 = large.shape[0]
        # Reduce height. No change to width.
        small = small[0:y2 - y1, 0:small.shape[1]]

    if x2 > large.shape[1]:
        x2 = large.shape[1]
        # No change to height. Reduce width.
        small = small[0:small.shape[0], 0:x2 - x1]

    alpha_s = small[:, :, 3] / 255.0
    alpha_l = 1.0 - alpha_s

    for c in range(0, 3):
        large[y1:y2, x1:x2, c] = (alpha_s * small[:, :, c] +
                                  alpha_l * large[y1:y2, x1:x2, c])
    return large


def image_paths(dir_or_filename):
    """Given:
     * a directory path,
     * a path/file specification,
     * or a single image,
    return a list of paths to all images
    """
    # Expand any "~"
    dir_or_filename = os.path.expanduser(dir_or_filename)

    if os.path.isdir(dir_or_filename):
        # Create a file spec
        dir_or_filename = os.path.join(dir_or_filename, "*")

    # Find all entries that match the spec
    filenames = glob.glob(dir_or_filename)

    # Filter out directories
    filenames = [f for f in filenames if os.path.isfile(f)]

    return filenames


def random_flip(image):
    """50% chance to flip the image for some variation"""
    if random.random() < 0.5:
        image = cv2.flip(image, 1)  # 1 = vertical flip
    return image


def show_image(image, show=False):
    if show:
        cv2.imshow('img', image)
        cv2.waitKey(0)
        cv2.destroyAllWindows()


def photobomb(infile, in_bodies, outfile, show=False):
    l_img = cv2.imread(infile)

    body_paths = image_paths(in_bodies)

    # Load a body, resize it and paste it
    random_index = random.randrange(0, len(body_paths))
    print(body_paths[random_index])

    s_img = cv2.imread(body_paths[random_index], -1)
    assert s_img is not None

    s_img = random_flip(s_img)

    # Resize photobomber to a fraction of full image width
    s_img = image_resize(s_img, width=int(l_img.shape[1] * 0.5))

    # And make sure photobomber is no taller than the full image
    if s_img.shape[0] > l_img.shape[0]:
        s_img = image_resize(s_img, height=int(l_img.shape[0] * 0.8))

    # Top-left paste coordinates
    x1 = random.randrange(0, l_img.shape[1] - s_img.shape[1])
    y1 = l_img.shape[0] - s_img.shape[0]

    # Bottom-right paste coordinates
    x2 = x1 + s_img.shape[1]
    y2 = y1 + s_img.shape[0]

    l_img = paste_image(l_img, s_img, x1, y1, x2, y2)

    show_image(l_img, show)

    cv2.imwrite(outfile, l_img)


def detect(infile, in_faces, outfile, face_cascade_path, eye_cascade_path,
           show=False, boxes=False):

    # A cache so we don't need to re-open the same image
    crisu_cache = {}

    face_paths = image_paths(in_faces)

    face_cascade = cv2.CascadeClassifier(face_cascade_path)
    eye_cascade = cv2.CascadeClassifier(eye_cascade_path)
    l_img = cv2.imread(infile)
    gray = cv2.cvtColor(l_img, cv2.COLOR_BGR2GRAY)

    faces = face_cascade.detectMultiScale(gray, 1.3, 5)
    print(faces)
    for (x, y, w, h) in faces:
        if boxes:
            cv2.rectangle(l_img, (x, y), (x+w, y+h), (255, 0, 0), 2)
            roi_gray = gray[y:y+h, x:x+w]
            roi_color = l_img[y:y+h, x:x+w]
            eyes = eye_cascade.detectMultiScale(roi_gray)
            for (ex, ey, ew, eh) in eyes:
                cv2.rectangle(
                    roi_color, (ex, ey), (ex+ew, ey+eh), (0, 255, 0), 2)

        # Load a Crisu head, resize it and paste it
        random_index = random.randrange(0, len(face_paths))
        print(face_paths[random_index])

        # Load from cache, or put in cache
        if random_index not in crisu_cache:
            crisu_cache[random_index] = cv2.imread(face_paths[random_index],
                                                   -1)
        s_img = crisu_cache[random_index]

        s_img = random_flip(s_img)

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

        if boxes:
            cv2.rectangle(l_img, (x, y), (x + w, y + h), (0, 0, 255), 2)

        # Bottom-right paste coordinates
        x2 = x + s_img.shape[1]
        y2 = y + s_img.shape[0]

        l_img = paste_image(l_img, s_img, x, y, x2, y2)

    if not len(faces):
        print("No faces detected")
        return False

    show_image(l_img, show)

    cv2.imwrite(outfile, l_img)
    return True


def check_path(path, filename):
    """Return full path if there's a file at the given path, else exit"""
    full_path = os.path.join(path, filename)
    if not os.path.isfile(full_path):
        sys.exit("File not found: {}".format(full_path))
    return full_path


if __name__ == '__main__':

    parser = argparse.ArgumentParser(
        description='Find faces in an input image and '
                    'replace them with other random faces',
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument(
        'infile',
        help='Image to mess up')
    parser.add_argument(
        '-f',
        '--faces',
        default='data/faces/*.png',
        help='Either a directory of images or a single face image')
    parser.add_argument(
        '-b',
        '--bodies',
        default='data/bodies/*.png',
        help='Either a directory of images or a single body image')
    parser.add_argument(
        '-o',
        '--outfile',
        default='out.jpg',
        help='Output image filename')
    parser.add_argument(
        '-cp',
        '--cascade-path',
        default='/usr/local/Cellar/opencv@2/2.4.13.5/share/OpenCV/'
                'haarcascades',
        help='Haar cascade file')
    parser.add_argument(
        '-fc',
        '--face-cascade',
        default='haarcascade_frontalface_alt.xml',
        help='Haar cascade file')
    parser.add_argument(
        '-ec',
        '--eye-cascade',
        default='haarcascade_eye.xml',
        help='Haar cascade file')
    parser.add_argument(
        '-p',
        '--photobomb',
        action='store_true',
        help='Photobomb instead of detecting')
    parser.add_argument(
        '-s',
        '--show',
        action='store_true',
        help='Debug: show output image in window')
    parser.add_argument(
        '-bx',
        '--boxes',
        action='store_true',
        help='Debug: draw boxes around deteted faces')

    args = parser.parse_args()

    detected = False
    if not args.photobomb:
        detected = detect(args.infile,
                          args.faces,
                          args.outfile,
                          check_path(args.cascade_path, args.face_cascade),
                          check_path(args.cascade_path, args.eye_cascade),
                          args.show,
                          args.boxes,
                          )

    if args.photobomb or not detected:
            photobomb(args.infile,
                      args.bodies,
                      args.outfile,
                      args.show,
                      )

# End of file
