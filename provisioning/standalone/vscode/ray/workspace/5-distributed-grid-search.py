"""
RAY DISTRIBUTED GRID SEARCH
---------------------------
This script demonstrates how to build a custom hyperparameter sweep 
using standard Ray Core (without Ray Tune).

It handles:
1. Generating a grid of hyperparameter combinations.
2. Submitting them as individual @ray.remote tasks.
3. Using a "Sliding Window" pattern to strictly limit concurrency 
   so we don't overwhelm the cluster queue.
4. Collecting and ranking the results.
"""

import ray
import time
import itertools
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, Dataset

# ==========================================
# ⚙️ CONFIGURATION SECTION
# ==========================================

RAY_HEAD_ADDRESS = "ray://raycluster-head-svc.workspace-kuberay-gpu.svc.cluster.local:10001"

# Grid Search Parameters (Add more values here to expand the sweep)
LEARNING_RATES = [0.1, 0.01, 0.001]
BATCH_SIZES    = [16, 32, 64]
HIDDEN_SIZES   = [32, 128]

# Execution Settings
MAX_CONCURRENT_TASKS = 4     # Limit how many tasks run at the exact same time
EPOCHS               = 15    # How long each config should train
USE_GPU              = True  # Set to True if your workers have GPUs to spare

# ==========================================


# --- 1. Basic Model & Data setup ---

class SimpleNN(nn.Module):
    def __init__(self, hidden_size):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(2, hidden_size),
            nn.ReLU(),
            nn.Linear(hidden_size, 2)
        )

    def forward(self, x):
        return self.net(x)

class DummyDataset(Dataset):
    def __init__(self, size=1000):
        self.x = torch.randn(size, 2)
        self.y = ((self.x[:, 0] + self.x[:, 1]) > 0).long()

    def __len__(self):
        return len(self.x)

    def __getitem__(self, idx):
        return self.x[idx], self.y[idx]


# --- 2. The Remote Training Task ---

# We explicitly request CPU/GPU resources here
@ray.remote(num_cpus=1, num_gpus=1 if USE_GPU else 0)
def evaluate_config(run_id: int, lr: float, batch_size: int, hidden_size: int) -> dict:
    """
    Trains a model from scratch using the provided configuration 
    and returns the final loss metric.
    """
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    
    # Setup model, optimizer, and data
    model = SimpleNN(hidden_size=hidden_size).to(device)
    optimizer = optim.SGD(model.parameters(), lr=lr)
    train_loader = DataLoader(DummyDataset(), batch_size=batch_size, shuffle=True)

    # Training Loop
    running_loss = 0.0
    for epoch in range(EPOCHS):
        model.train()
        running_loss = 0.0
        
        for x, y in train_loader:
            x, y = x.to(device), y.to(device)
            optimizer.zero_grad()
            output = model(x)
            loss = nn.CrossEntropyLoss()(output, y)
            loss.backward()
            optimizer.step()
            running_loss += loss.item()

    # Calculate final average loss for the last epoch
    final_avg_loss = running_loss / len(train_loader)
    
    # Return the metrics so the driver script can rank them
    return {
        "run_id": run_id,
        "lr": lr,
        "batch_size": batch_size,
        "hidden_size": hidden_size,
        "loss": final_avg_loss
    }


# --- 3. Cluster Execution & Sliding Window ---

def main():
    print(f"Step 1: Connecting to Ray cluster at {RAY_HEAD_ADDRESS}...")
    
    try:
        ray.init(
            address=RAY_HEAD_ADDRESS,
            runtime_env={"pip": ["torch==2.4.1"]},
            ignore_reinit_error=True
        )
        print("✅ Connected to cluster!\n")

        # Create all possible combinations of our hyperparameters
        config_grid = list(itertools.product(LEARNING_RATES, BATCH_SIZES, HIDDEN_SIZES))
        total_runs = len(config_grid)
        
        print(f"Step 2: Starting Custom Grid Search ({total_runs} combinations)...")
        print(f"Limiting execution to {MAX_CONCURRENT_TASKS} parallel tasks at a time.\n")

        start_time = time.time()
        
        in_flight_futures = []
        completed_results = []
        
        # --- The Sliding Window Pattern ---
        for run_id, (lr, batch_size, hidden_size) in enumerate(config_grid):
            
            # If we hit our concurrency limit, wait for at least 1 task to finish
            if len(in_flight_futures) >= MAX_CONCURRENT_TASKS:
                # ray.wait returns two lists: ready tasks and tasks still pending
                ready, in_flight_futures = ray.wait(in_flight_futures, num_returns=1)
                
                # Fetch the actual data from the ready task and store it
                finished_data = ray.get(ready[0])
                completed_results.append(finished_data)
                print(f"  ✓ Task {finished_data['run_id']:02d} finished | Loss: {finished_data['loss']:.4f}")

            # Submit the next task
            future = evaluate_config.remote(run_id, lr, batch_size, hidden_size)
            in_flight_futures.append(future)

        # Wait for the remaining tasks in the queue to finish
        if in_flight_futures:
            print("\n⏳ Waiting for the final tasks to complete...")
            remaining_results = ray.get(in_flight_futures)
            completed_results.extend(remaining_results)
            for res in remaining_results:
                print(f"  ✓ Task {res['run_id']:02d} finished | Loss: {res['loss']:.4f}")

        # --- Ranking the Results ---
        elapsed = time.time() - start_time
        
        # Sort results by lowest loss
        completed_results.sort(key=lambda x: x["loss"])
        best = completed_results[0]

        print(f"\nStep 3: Sweep Complete in {elapsed:.1f}s! Analyzing results...")
        
        print("\n=== 🏆 BEST CONFIGURATION FOUND ===")
        print(f"Run ID        : {best['run_id']:02d}")
        print(f"Learning Rate : {best['lr']}")
        print(f"Batch Size    : {best['batch_size']}")
        print(f"Hidden Size   : {best['hidden_size']}")
        print(f"Final Loss    : {best['loss']:.4f}")
        print("===================================\n")

        print("--- Top 3 Runners Up ---")
        for i in range(1, min(4, len(completed_results))):
            res = completed_results[i]
            print(f" #{i+1} -> LR: {res['lr']}, Batch: {res['batch_size']}, "
                  f"Hidden: {res['hidden_size']} | Loss: {res['loss']:.4f}")

    except Exception as e:
        print(f"\n❌ Script execution error: {e}")
    finally:
        if ray.is_initialized():
            ray.shutdown()
            print("\nRay connection closed. Resources released.")

if __name__ == "__main__":
    main()