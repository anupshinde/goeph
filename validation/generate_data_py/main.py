import pandas as pd
import celestial
from tqdm import tqdm
import os
import pytz
from datetime import datetime, timedelta


def create_batch_processed_planetary_df(input_df, output_file_name):
    batch_size = 100000
    num_batches = len(input_df) // batch_size + (1 if len(input_df) % batch_size != 0 else 0)
    output_file = output_file_name
    os.remove(output_file) if os.path.exists(output_file) else None
    for i in tqdm(range(num_batches)):
        batch_df = input_df.iloc[i * batch_size:(i + 1) * batch_size]
        result_df = celestial.get_celestial_df(batch_df["Time"], addSatellites=True)

        result_df.to_csv(output_file, mode='a', header=not i, index=False)

def get_dates_df(startDate, endDate):
    startDateStr = startDate.date().isoformat()
    endDateStr = endDate.date().isoformat()
    df = pd.DataFrame()
    df["Time"] = pd.date_range(start=startDateStr, end=endDateStr, freq='1h',tz='UTC')
    return df

def main():
    referenceDate = datetime(2026, 1, 19, tzinfo=pytz.UTC)
    pastDate = referenceDate + timedelta(days=-100*365) # ±100 years
    futureDate = referenceDate + timedelta(days=100*365) # ±100 years
    
    df = get_dates_df(pastDate, futureDate)
    print(df)

    create_batch_processed_planetary_df(df,"/tmp/planetary_data_py.csv")

if __name__ == "__main__":
    main()
