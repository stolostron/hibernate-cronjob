FROM registry.access.redhat.com/ubi8/python-38

# Add the program and the configuration file
ADD ./hibernate-cronjob/action.py .
ADD ./hibernate-cronjob/event.py .

# Install dependencies
RUN pip install --upgrade pip
RUN python -m pip install kubernetes

USER 1001