"""
RAY DEEP NEURAL NETWORK TRAINING
--------------------------------
This script demonstrates how to train a deep neural network 
on a Ray cluster using a dedicated GPU worker. 

It handles:
1. Connecting to the Ray cluster.
2. Generating a synthetic dataset directly on the worker.
3. Training a Deep PyTorch model using the specified hyperparameters.
4. Returning the trained model weights to the driver.
5. Saving the model to a shared storage path (NFS/Persistent Volume).
"""

import os
import time
import torch
import torch.nn.functional as F
from torch.utils.data import Dataset, DataLoader
import ray

# ==========================================
# ⚙️ CONFIGURATION SECTION
# Modify these variables to adjust the run
# ==========================================

# Cluster & Storage Settings
RAY_HEAD_ADDRESS    = "ray://raycluster-head-svc.workspace-kuberay-gpu.svc.cluster.local:10001"
SHARED_STORAGE_PATH = "/shared/models/neural_network_ray_example.pth"

# Training Hyperparameters
NUM_EPOCHS    = 100      # Number of times to iterate over the dataset
BATCH_SIZE    = 64       # Number of samples per gradient update
LEARNING_RATE = 0.01     # Step size for the optimizer
HIDDEN_SIZE   = 256      # Number of neurons in the primary hidden layers
NUM_SAMPLES   = 5000     # Size of the synthetic dataset to generate
RANDOM_SEED   = 42       # For reproducibility

# ==========================================


class NeuralNetwork(torch.nn.Module):
    """
    A deep multi-layer perceptron (MLP) with extended steps.
    Gradually funnels the features down through multiple hidden layers.
    """
    def __init__(self, num_inputs: int, num_outputs: int, hidden_size: int):
        super().__init__()
        
        self.layers = torch.nn.Sequential(
            # Input layer expanding to hidden_size
            torch.nn.Linear(num_inputs, hidden_size),
            torch.nn.ReLU(),
            torch.nn.Dropout(0.2),

            # Step 1: Maintain width
            torch.nn.Linear(hidden_size, hidden_size),
            torch.nn.ReLU(),
            torch.nn.Dropout(0.2),

            # Step 2: Maintain width
            torch.nn.Linear(hidden_size, hidden_size),
            torch.nn.ReLU(),
            torch.nn.Dropout(0.2),

            # Step 3: Start funneling down (Half size)
            torch.nn.Linear(hidden_size, hidden_size // 2),
            torch.nn.ReLU(),
            torch.nn.Dropout(0.2),

            # Step 4: Maintain half size
            torch.nn.Linear(hidden_size // 2, hidden_size // 2),
            torch.nn.ReLU(),
            torch.nn.Dropout(0.2),

            # Step 5: Quarter size
            torch.nn.Linear(hidden_size // 2, hidden_size // 4),
            torch.nn.ReLU(),
            torch.nn.Dropout(0.2),

            # Step 6: Eighth size
            torch.nn.Linear(hidden_size // 4, hidden_size // 8),
            torch.nn.ReLU(),

            # Output layer mapping to final classes
            torch.nn.Linear(hidden_size // 8, num_outputs),
        )

    def forward(self, x):
        return self.layers(x)


class ToyDataset(Dataset):
    """Simple dataset wrapper for our PyTorch DataLoader."""
    def __init__(self, X, y):
        self.features = X
        self.labels = y

    def __getitem__(self, index):
        return self.features[index], self.labels[index]

    def __len__(self):
        return self.labels.shape[0]


def compute_accuracy(model, dataloader, device):
    """Helper to calculate the accuracy of the model."""
    model.eval()
    correct = 0.0
    total = 0

    with torch.no_grad():
        for features, labels in dataloader:
            features, labels = features.to(device), labels.to(device)
            logits = model(features)
            predictions = torch.argmax(logits, dim=1)
            correct += torch.sum(labels == predictions).item()
            total += len(labels)

    return correct / total


@ray.remote(num_gpus=1)
def train_model_task() -> dict:
    """
    This function runs entirely on the remote Ray GPU worker.
    It generates data, trains the deep model, and returns the CPU-safe weights.
    """
    # 1. Environment Setup
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    worker_id = ray.get_runtime_context().get_worker_id()
    
    print(f"\n🚀 [Worker {worker_id[:8]}] Initialized on device: {device}")
    torch.manual_seed(RANDOM_SEED)

    # 2. Data Generation (Synthetic 2D classification problem)
    print(f"📊 [Worker {worker_id[:8]}] Generating {NUM_SAMPLES} samples...")
    X_train = torch.randn(NUM_SAMPLES, 2).to(device)
    # Target: 1 if sum of coordinates > 0, else 0
    y_train = ((X_train[:, 0] + X_train[:, 1]) > 0).long().to(device)

    test_size = NUM_SAMPLES // 5
    X_test = torch.randn(test_size, 2).to(device)
    y_test = ((X_test[:, 0] + X_test[:, 1]) > 0).long().to(device)

    train_loader = DataLoader(ToyDataset(X_train.cpu(), y_train.cpu()), batch_size=BATCH_SIZE, shuffle=True)
    test_loader  = DataLoader(ToyDataset(X_test.cpu(), y_test.cpu()), batch_size=BATCH_SIZE, shuffle=False)

    # 3. Model & Optimizer Initialization
    model = NeuralNetwork(num_inputs=2, num_outputs=2, hidden_size=HIDDEN_SIZE).to(device)
    optimizer = torch.optim.SGD(model.parameters(), lr=LEARNING_RATE)

    # 4. Training Loop
    print(f"⚙️  [Worker {worker_id[:8]}] Starting training ({NUM_EPOCHS} Epochs)...")
    start_time = time.time()
    
    for epoch in range(NUM_EPOCHS):
        model.train()
        epoch_loss = 0.0
        
        for features, labels in train_loader:
            features, labels = features.to(device), labels.to(device)
            
            # Forward & Backward pass
            optimizer.zero_grad()
            logits = model(features)
            loss = F.cross_entropy(logits, labels)
            loss.backward()
            optimizer.step()
            
            epoch_loss += loss.item()
            
        # Print progress every 10 epochs or on the last epoch
        if (epoch + 1) % 10 == 0 or epoch == NUM_EPOCHS - 1:
            avg_loss = epoch_loss / len(train_loader)
            print(f"   🔄 Epoch {epoch+1:03d}/{NUM_EPOCHS:03d} | Avg Loss: {avg_loss:.4f}")

    training_time = time.time() - start_time

    # 5. Evaluation
    train_acc = compute_accuracy(model, train_loader, device)
    test_acc  = compute_accuracy(model, test_loader, device)
    
    print(f"✅ [Worker {worker_id[:8]}] Training complete in {training_time:.1f}s")
    print(f"🎯 [Worker {worker_id[:8]}] Final Test Accuracy: {test_acc:.4f}")

    # 6. Safe Serialization
    # CRITICAL: We must move the model to CPU before returning its state_dict. 
    # If we return CUDA tensors, Ray will crash trying to serialize them back to the head node.
    model_cpu = model.cpu()
    state_dict_cpu = model_cpu.state_dict()

    if torch.cuda.is_available():
        torch.cuda.empty_cache()

    return {
        "success": True,
        "training_time": training_time,
        "test_accuracy": test_acc,
        "model_state_dict": state_dict_cpu
    }


# Notice: num_cpus=1 instead of num_gpus=1. 
# Saving a file to disk doesn't require a GPU, so we don't waste expensive resources here.
@ray.remote(num_cpus=1)
def save_model_to_disk(state_dict: dict, file_path: str) -> bool:
    """Saves the trained PyTorch state dictionary to the shared cluster storage."""
    try:
        # Ensure the directory exists
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        # Save the weights
        torch.save(state_dict, file_path)
        return True
    except Exception as e:
        print(f"❌ Error saving model to {file_path}: {e}")
        return False


def main():
    print(f"Step 1: Connecting to Ray cluster at {RAY_HEAD_ADDRESS}...")
    
    try:
        # We ensure PyTorch is installed on the worker before starting
        ray.init(
            address=RAY_HEAD_ADDRESS,
            runtime_env={"pip": ["torch==2.4.1"]},
            ignore_reinit_error=True
        )
        print("✅ Connected to cluster!\n")

        # Display active configuration
        print("--- Training Configuration ---")
        print(f"Epochs       : {NUM_EPOCHS}")
        print(f"Batch Size   : {BATCH_SIZE}")
        print(f"Learning Rate: {LEARNING_RATE}")
        print(f"Hidden Size  : {HIDDEN_SIZE}")
        print("------------------------------\n")

        print("Step 2: Submitting Deep GPU training task to cluster...")
        # Submit the training task
        future = train_model_task.remote()
        
        # Block until training is finished and get the result
        result = ray.get(future)

        if not result["success"]:
            print("❌ Training failed on the worker.")
            return

        print("\nStep 3: Saving model to shared storage...")
        # Submit the saving task. We pass the state_dict from the training result.
        save_future = save_model_to_disk.remote(result["model_state_dict"], SHARED_STORAGE_PATH)
        saved_successfully = ray.get(save_future)

        if saved_successfully:
            print(f"✅ Model successfully saved to: {SHARED_STORAGE_PATH}")
        else:
            print("❌ Failed to save model.")

    except Exception as e:
        print(f"\n❌ Script execution error: {e}")
    finally:
        if ray.is_initialized():
            ray.shutdown()
            print("\nRay connection closed. Resources released.")

if __name__ == "__main__":
    main()