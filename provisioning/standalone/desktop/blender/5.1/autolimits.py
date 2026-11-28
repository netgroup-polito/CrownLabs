import os
import bpy
from bpy.app.handlers import persistent
import math


@persistent  # Keep across file reloads
def cpulimits_setter(_):  # An argument has to be present for handlers
    try:
        maxThreads = math.ceil(float(os.environ['CROWNLABS_CPU_LIMITS']))  # Round possible floats
        # There might be more than a scene, we cycle them all
        for scn in bpy.data.scenes:
            scn.render.threads_mode = 'FIXED'
            scn.render.threads = maxThreads
    except Exception as e:
        print("Error while setting rendering threads", e)


def register():  # Required method, called on startup
    # Handlers are called on certain events,
    # load_post occurs after a file has been loaded
    bpy.app.handlers.load_post.append(cpulimits_setter)
