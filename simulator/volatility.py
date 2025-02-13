# %%
import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

def load_data():
    # Parameters
    fee_rate = 0.00055
    DipsRatio = -0.01
    TargetTimeForMinutes = 1

    # Setup data directory and list files with .csv.gz extension
    data_dir = os.path.join(os.getcwd(), 'data')
    files = [f for f in os.listdir(data_dir) if f.endswith('.csv.gz')]
    files.sort()

    # Read ohlcv data from each file
    columns = ['timestamp', 'open', 'high', 'low', 'close', 'volume']
    ohlcvs_list = []
    for file in files:
        temp = pd.read_csv(os.path.join(data_dir, file))
        temp.columns = columns
        ohlcvs_list.append(temp)
    # Concatenate into a single pandas DataFrame
    ohlcvs = pd.concat(ohlcvs_list, ignore_index=True)
    return ohlcvs, DipsRatio, TargetTimeForMinutes

def calculate_volatility(df, TargetTimeForMinutes):
    # Calculate shifted data for volatility and target_y columns
    shift = df.shift(-TargetTimeForMinutes)
    target = df.shift(-(TargetTimeForMinutes + 1))
    df['volatility'] = (shift['close'] - df['close']) / df['close']
    df['target_y'] = (target['close'] - shift['close']) / shift['close']
    return df

def plot_histogram(df):
    print(df['volatility'].describe())
    hist_target = 0.01
    filtered = df[(df['volatility'] < hist_target) & (df['volatility'] > -hist_target)]
    plt.hist(filtered['volatility'], bins=100)
    plt.title('Histogram of Volatility')
    plt.show()

def plot_scatter_with_regression(df, DipsRatio):
    # Filter the data based on the dips ratio flag
    df['flag'] = df['volatility'] < DipsRatio
    filtered = df[df['flag']]
    
    plt.figure(figsize=(12, 6))
    plt.scatter(x='volatility', y='target_y', data=filtered, alpha=0.5)
    
    # Add linear regression if there are enough data points
    mask = ~(filtered['volatility'].isna() | filtered['target_y'].isna())
    x = filtered.loc[mask, 'volatility']
    y = filtered.loc[mask, 'target_y']
    if len(x) > 1:
        z = np.polyfit(x, y, 1)
        p = np.poly1d(z)
        xs = np.sort(x)
        plt.plot(xs, p(xs), "r--", alpha=0.8)
        plt.text(0.05, 0.95, f'Slope: {z[0]:.4f}', transform=plt.gca().transAxes)
        plt.text(0.05, 0.90, f'Intercept: {z[1]:.4f}', transform=plt.gca().transAxes)

    plt.xlabel('Volatility')
    plt.ylabel('Target Y')
    plt.title('Volatility vs Target Y with Linear Regression')
    plt.show()

def main():
    # Load and display initial data
    ohlcvs, DipsRatio, TargetTimeForMinutes = load_data()
    print("First row of the data:")
    print(ohlcvs.iloc[0])
    print("Last row of the data:")
    print(ohlcvs.iloc[-1])
    
    # Calculate volatility metrics
    ohlcvs = calculate_volatility(ohlcvs, TargetTimeForMinutes)
    
    # Plot histogram and scatter plots
    plot_histogram(ohlcvs)
    plot_scatter_with_regression(ohlcvs, DipsRatio)

if __name__ == '__main__':
    main()


# %%



