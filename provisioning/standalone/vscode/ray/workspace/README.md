# 🚀 Ray Cluster on Crownlabs: Quickstart Guide

Welcome to the **Crownlabs Ray Cluster**.
This guide provides everything you need to start distributing your Python and Machine Learning workloads across our infrastructure. 

Ray is a powerful distributed computing framework.
Instead of rewriting your code from scratch, Ray allows you to take standard Python functions and scale them across multiple CPU and GPU worker nodes with just a few lines of code.

---

## 🔌 Connecting to the Cluster
Since Ray and PyTorch are already installed and configured in your environment, you can jump straight into the code! 

To execute code on the cluster, your Python script must connect to the Ray Head Node.
On Crownlabs, the internal cluster address is:

**Address:** `ray://raycluster-head-svc.workspace-kuberay-gpu.svc.cluster.local:10001`

You initialize this connection at the very top of your script using `ray.init()`:

```python
import ray

RAY_HEAD = "ray://raycluster-head-svc.workspace-kuberay-gpu.svc.cluster.local:10001"

# Connect to the cluster
ray.init(
    address=RAY_HEAD,
    ignore_reinit_error=True
)
```

---

## 🧠 The 3 Core Concepts
Using Ray boils down to three simple steps.

### 1. The Decorator (`@ray.remote`)
To tell Ray that a function should run on the cluster (instead of your local machine), add the `@ray.remote` decorator.
You can also specify exactly what hardware the function needs.

```python
# This task requires exactly 1 GPU to run
@ray.remote(num_gpus=1)
def train_model():
    import torch
    return torch.cuda.is_available()

# This task only needs a CPU core
@ray.remote(num_cpus=1)
def calculate_math():
    return 1 + 1
```

### 2. Submitting Tasks (`.remote()`)
To execute the function on the cluster, do not call it normally (e.g., `train_model()`).
Instead, append `.remote()`.
This immediately sends the task to the cluster queue and returns a "Future" (an Object Reference), allowing your main script to keep running without freezing.

```python
# Submit 10 tasks in parallel
futures = [calculate_math.remote() for _ in range(10)]
```

### 3. Retrieving Results (`ray.get()`)
To get the actual data back from the cluster, pass the Future(s) to `ray.get()`.
**Note:** This is a blocking operation; your script will pause here until the cluster finishes the work.

```python
results = ray.get(futures)
print(results) # Output: [2, 2, 2, 2...]
```

---

## 📂 Sample Scripts
Inside the `samples/` folder, you will find ready-to-run scripts progressing from basic sanity checks to advanced machine learning sweeps.
We highly recommend running them in this order:

| Script | Purpose | Description |
| :--- | :--- | :--- |
| `cpu_sum_test.py` | CPU Diagnostics | Verifies that your code can connect to the cluster and distribute standard Python math tasks across CPU cores. |
| `gpu_sanity_check.py` | GPU Diagnostics | A quick check to verify NVIDIA drivers are working and PyTorch can utilize the GPUs. |
| `matrix_multiply.py` | Performance Benchmark | A heavy stress-test that distributes massive matrix calculations to measure the cluster's aggregate TFLOPS. |
| `single_gpu_training.py` | Deep Learning | A complete template for generating data, training a PyTorch Neural Network, and saving the model to shared storage. |
| `grid_search_sweep.py` | Hyperparameter Tuning | Advanced script showing how to train multiple models in parallel with strict concurrency limits. |

---

## ⚠️ Best Practices & Gotchas

- **Returning Tensors:** Never return a CUDA (GPU) tensor back to the driver script. Ray cannot serialize GPU memory over the network. Always move models and tensors to CPU (`model.cpu()`) before returning them in your remote functions.
- **Imports inside Tasks:** Because your functions execute on remote machines, any libraries you need (like `import torch`) must be imported **inside** the `def my_function():` block.
- **Clean Shutdown:** Always put `ray.shutdown()` at the end of your scripts (ideally in a `finally` block) to politely disconnect and release cluster resources for other users.